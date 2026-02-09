package app

import (
	tea "charm.land/bubbletea/v2"
	"context"
	"errors"
	"fmt"
	"github.com/charmbracelet/catwalk/pkg/catwalk"
	"github.com/prometheus/common/version"
	"github.com/xiehqing/common/agent/agent"
	"github.com/xiehqing/common/agent/config"
	"github.com/xiehqing/common/agent/csync"
	"github.com/xiehqing/common/agent/history"
	"github.com/xiehqing/common/agent/lsp"
	"github.com/xiehqing/common/agent/mcp"
	"github.com/xiehqing/common/agent/message"
	"github.com/xiehqing/common/agent/permission"
	"github.com/xiehqing/common/agent/pubsub"
	"github.com/xiehqing/common/agent/session"
	"github.com/xiehqing/common/agent/shell"
	"github.com/xiehqing/common/agent/update"
	"github.com/xiehqing/common/pkg/logs"
	"io"
	"sync"
	"time"
)

type App struct {
	Sessions         session.Service
	Messages         message.Service
	History          history.Service
	Permissions      permission.Service
	config           *config.Config
	AgentCoordinator agent.Coordinator
	LSPClients       *csync.Map[string, *lsp.Client]
	serviceEventsWG  *sync.WaitGroup
	eventsCtx        context.Context
	events           chan tea.Msg
	globalCtx        context.Context
	cleanupFuncs     []func() error
}

func New(ctx context.Context, additionalSystemPrompt string, cfg *config.Config, sessions session.Service, messages message.Service, files history.Service) (*App, error) {
	skipPermissionsRequests := cfg.Permissions != nil && cfg.Permissions.SkipRequests
	var allowedTools []string
	if cfg.Permissions != nil && cfg.Permissions.AllowedTools != nil {
		allowedTools = cfg.Permissions.AllowedTools
	}
	app := &App{
		Sessions:        sessions,
		Messages:        messages,
		History:         files,
		Permissions:     permission.NewPermissionService(cfg.WorkingDir(), skipPermissionsRequests, allowedTools),
		LSPClients:      csync.NewMap[string, *lsp.Client](),
		globalCtx:       ctx,
		config:          cfg,
		events:          make(chan tea.Msg, 100),
		serviceEventsWG: &sync.WaitGroup{},
	}
	app.setupEvents()
	// Check for updates in the background.
	go app.checkForUpdates(ctx)

	go func() {
		logs.Infof("Initializing MCP clients")
		mcp.Initialize(ctx, app.Permissions, cfg)
	}()

	// cleanup database upon app shutdown
	app.cleanupFuncs = append(app.cleanupFuncs, mcp.Close)

	// TODO: remove the concept of agent config, most likely.
	if !cfg.IsConfigured() {
		logs.Warnf("No agent configuration found")
		return app, nil
	}

	if err := app.InitCoderAgent(ctx, additionalSystemPrompt); err != nil {
		return nil, fmt.Errorf("failed to initialize coder agent: %w", err)
	}
	return app, nil
}

func (app *App) setupEvents() {
	ctx, cancel := context.WithCancel(app.globalCtx)
	app.eventsCtx = ctx
	setupSubscriber(ctx, app.serviceEventsWG, "sessions", app.Sessions.Subscribe, app.events)
	setupSubscriber(ctx, app.serviceEventsWG, "messages", app.Messages.Subscribe, app.events)
	setupSubscriber(ctx, app.serviceEventsWG, "permissions", app.Permissions.Subscribe, app.events)
	setupSubscriber(ctx, app.serviceEventsWG, "permissions-notifications", app.Permissions.SubscribeNotifications, app.events)
	setupSubscriber(ctx, app.serviceEventsWG, "history", app.History.Subscribe, app.events)
	//setupSubscriber(ctx, app.serviceEventsWG, "mcp", mcp.SubscribeEvents, app.events)
	//setupSubscriber(ctx, app.serviceEventsWG, "lsp", SubscribeLSPEvents, app.events)
	cleanupFunc := func() error {
		cancel()
		app.serviceEventsWG.Wait()
		return nil
	}
	app.cleanupFuncs = append(app.cleanupFuncs, cleanupFunc)
}

func (app *App) InitCoderAgent(ctx context.Context, additionalSystemPrompt string) error {
	coderAgentCfg := app.config.Agents[config.AgentCoder]
	if coderAgentCfg.ID == "" {
		return fmt.Errorf("coder agent configuration is missing")
	}
	var err error
	app.AgentCoordinator, err = agent.NewCoordinator(
		ctx,
		app.config,
		additionalSystemPrompt,
		app.Sessions,
		app.Messages,
		app.Permissions,
		app.History,
		app.LSPClients,
	)
	if err != nil {
		logs.Errorf("Failed to create coder agent，error：%v", err)
		return err
	}
	return nil
}

// GetDefaultSmallModel returns the default small model for the given
// provider. Falls back to the large model if no default is found.
func (app *App) GetDefaultSmallModel(providerID string) config.SelectedModel {
	cfg := app.config
	largeModelCfg := cfg.Models[config.SelectedModelTypeLarge]

	// Find the provider in the known provider list to get its default small model.
	knownProviders, _ := config.Providers(cfg)
	openProviders, err := config.OpenProviders(cfg)
	if err == nil {
		knownProviders = append(knownProviders, openProviders...)
	}
	var knownProvider *catwalk.Provider
	for _, p := range knownProviders {
		if string(p.ID) == providerID {
			knownProvider = &p
			break
		}
	}

	// For unknown/local provider, use the large model as small.
	if knownProvider == nil {
		logs.Warnf("Using large model as small model for unknown provider, provider:%s, model:%s", providerID, largeModelCfg.Model)
		return largeModelCfg
	}

	defaultSmallModelID := knownProvider.DefaultSmallModelID
	model := cfg.GetModel(providerID, defaultSmallModelID)
	if model == nil {
		logs.Warnf("Default small model not found, using large model, provider:%s, model:%s", providerID, largeModelCfg.Model)
		return largeModelCfg
	}

	logs.Warnf("Using provider default small model, provider:%s, model:%s", providerID, defaultSmallModelID)
	return config.SelectedModel{
		Provider:        providerID,
		Model:           defaultSmallModelID,
		MaxTokens:       model.DefaultMaxTokens,
		ReasoningEffort: model.DefaultReasoningEffort,
	}
}

func setupSubscriber[T any](
	ctx context.Context,
	wg *sync.WaitGroup,
	name string,
	subscriber func(context.Context) <-chan pubsub.Event[T],
	outputCh chan<- tea.Msg,
) {
	wg.Go(func() {
		subCh := subscriber(ctx)
		for {
			select {
			case event, ok := <-subCh:
				if !ok {
					logs.Debugf("subscription channel closed, name: %s", name)
					return
				}
				var msg tea.Msg = event
				select {
				case outputCh <- msg:
				case <-time.After(10 * time.Second):
					logs.Warnf("message dropped due to slow consumer, name: %s", name)
				case <-ctx.Done():
					logs.Debugf("subscriber cancelled, name: %s", name)
					return
				}
			case <-ctx.Done():
				logs.Debugf("subscriber cancelled, name: %s", name)
				return
			}
		}
	})
}

// checkForUpdates checks for available updates.
func (app *App) checkForUpdates(ctx context.Context) {
	checkCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	info, err := update.Check(checkCtx, version.Version, update.Default)
	if err != nil || !info.Available() {
		return
	}
	app.events <- UpdateAvailableMsg{
		CurrentVersion: info.Current,
		LatestVersion:  info.Latest,
		IsDevelopment:  info.IsDevelopment(),
	}
}

// UpdateAvailableMsg is sent when a new version is available.
type UpdateAvailableMsg struct {
	CurrentVersion string
	LatestVersion  string
	IsDevelopment  bool
}

// Shutdown performs a graceful shutdown of the application.
func (app *App) Shutdown() {
	start := time.Now()
	defer func() {
		logs.Infof("Shutdown took " + time.Since(start).String())
	}()

	// First, cancel all agents and wait for them to finish. This must complete
	// before closing the DB so agents can finish writing their state.
	if app.AgentCoordinator != nil {
		app.AgentCoordinator.CancelAll()
	}

	// Now run remaining cleanup tasks in parallel.
	var wg sync.WaitGroup

	// Kill all background shells.
	wg.Go(func() {
		shell.GetBackgroundShellManager().KillAll()
	})

	// Shutdown all LSP clients.
	shutdownCtx, cancel := context.WithTimeout(app.globalCtx, 5*time.Second)
	defer cancel()
	for name, client := range app.LSPClients.Seq2() {
		wg.Go(func() {
			if err := client.Close(shutdownCtx); err != nil &&
				!errors.Is(err, io.EOF) &&
				!errors.Is(err, context.Canceled) &&
				err.Error() != "signal: killed" {
				logs.Warnf("Failed to shutdown LSP client, name：%s, error：%v", name, err)
			}
		})
	}

	// Call all cleanup functions.
	for _, cleanup := range app.cleanupFuncs {
		if cleanup != nil {
			wg.Go(func() {
				if err := cleanup(); err != nil {
					logs.Errorf("Failed to cleanup app properly on shutdown，error：%v", err)
				}
			})
		}
	}
	wg.Wait()
}
