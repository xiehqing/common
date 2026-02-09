package config

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/charmbracelet/catwalk/pkg/catwalk"
	"github.com/charmbracelet/catwalk/pkg/embedded"
	"github.com/xiehqing/common/agent/csync"
	"github.com/xiehqing/common/agent/home"
	"github.com/xiehqing/common/pkg/logs"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
	"time"
)

type syncer[T any] interface {
	Get(context.Context) (T, error)
}

var (
	providerOnce sync.Once
	providerList []catwalk.Provider
	providerErr  error
)

var (
	catwalkSyncer = &catwalkSync{}
	//hyperSyncer   = &hyperSync{}
)

// UpdateProviders 更新提供者列表
func UpdateProviders(pathOrUrl string) error {
	var providers []catwalk.Provider
	pathOrUrl = cmp.Or(pathOrUrl, os.Getenv(EnvCatwalkUrl), defaultCatwalkURL)
	switch {
	case pathOrUrl == "embedded":
		providers = embedded.GetAll()
	case strings.HasPrefix(pathOrUrl, "http://") || strings.HasPrefix(pathOrUrl, "https://"):
		var err error
		providers, err = catwalk.NewWithURL(pathOrUrl).GetProviders(context.Background(), "")
		if err != nil {
			return fmt.Errorf("failed to fetch provider from %s: %w", pathOrUrl, err)
		}
	default:
		content, err := os.ReadFile(pathOrUrl)
		if err != nil {
			return fmt.Errorf("failed to read provider file: %w", err)
		}
		if err := json.Unmarshal(content, &providers); err != nil {
			return fmt.Errorf("failed to unmarshal provider data: %w", err)
		}
		if len(providers) == 0 {
			return fmt.Errorf("no provider found in the provided source")
		}
	}
	if err := newCache[[]catwalk.Provider](cachePathFor("providers")).Store(providers); err != nil {
		return fmt.Errorf("failed to save provider to cache: %w", err)
	}
	logs.Infof("Providers updated successfully, count: %d, from: %s, to: %s", len(providers), pathOrUrl, cachePathFor)
	return nil
}

// Providers 返回提供者列表，同时考虑缓存结果以及是否启用了自动更新。
//
// 具体流程如下：
//
// 如果自动更新被禁用，它将返回发布时内置的提供者列表。
// 加载缓存的提供者列表。
// 尝试获取最新的提供者列表，然后根据情况返回：新的列表、缓存的列表，或者如果其他方式都失败，则返回内置列表。
func Providers(cfg *Config) ([]catwalk.Provider, error) {
	providerOnce.Do(func() {
		var wg sync.WaitGroup
		var errs []error
		providers := csync.NewSlice[catwalk.Provider]()
		autoUpdate := !cfg.Options.DisableProviderAutoUpdate
		ctx, cancelFunc := context.WithTimeout(context.Background(), 45*time.Second)
		defer cancelFunc()
		wg.Go(func() {
			catwalkUrl := cmp.Or(os.Getenv(EnvCatwalkUrl), defaultCatwalkURL)
			logs.Infof("Fetching Catwalk provider from %s", catwalkUrl)
			client := catwalk.NewWithURL(catwalkUrl)
			path := cachePathFor("providers")
			catwalkSyncer.Init(client, path, autoUpdate)
			items, err := catwalkSyncer.Get(ctx)
			if err != nil {
				catwalkUrl := fmt.Sprintf("%s/v2/provider", cmp.Or(os.Getenv(EnvCatwalkUrl), defaultCatwalkURL))
				errs = append(errs, fmt.Errorf("Crush was unable to fetch an updated list of provider from %s. Consider setting CRUSH_DISABLE_PROVIDER_AUTO_UPDATE=1 to use the embedded provider bundled at the time of this Crush release. You can also update provider manually. For more info see crush update-provider --help.\n\nCause: %w", catwalkUrl, providerErr)) //nolint:staticcheck
				return
			}
			providers.Append(items...)
		})
		wg.Wait()
		providerList = slices.Collect(providers.Seq())
		providerErr = errors.Join(errs...)
	})
	return providerList, providerErr
}

// cachePathFor 返回缓存文件路径
func cachePathFor(name string) string {
	xdgDataHome := os.Getenv(EnvXdgDataHome)
	if xdgDataHome != "" {
		return filepath.Join(xdgDataHome, appName, name+".json")
	}

	// return the path to the main data directory
	// for windows, it should be in `%LOCALAPPDATA%/crush/`
	// for linux and macOS, it should be in `$HOME/.local/share/crush/`
	if runtime.GOOS == "windows" {
		localAppData := os.Getenv(EnvLocalAppData)
		if localAppData != "" {
			localAppData = filepath.Join(os.Getenv(EnvUserProfile), "AppData", "Local")
		}
		return filepath.Join(localAppData, appName, name+".json")
	}

	return filepath.Join(home.Dir(), ".local", "share", appName, name+".json")
}
