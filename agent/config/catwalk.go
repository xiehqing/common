package config

import (
	"context"
	"errors"
	"github.com/charmbracelet/catwalk/pkg/catwalk"
	"github.com/charmbracelet/catwalk/pkg/embedded"
	"github.com/xiehqing/common/pkg/logs"
	"sync"
	"sync/atomic"
)

type catwalkClient interface {
	GetProviders(context.Context, string) ([]catwalk.Provider, error)
}

var _ syncer[[]catwalk.Provider] = (*catwalkSync)(nil)

type catwalkSync struct {
	once       sync.Once
	result     []catwalk.Provider
	cache      cache[[]catwalk.Provider]
	client     catwalkClient
	autoUpdate bool
	init       atomic.Bool
}

func (cs *catwalkSync) Init(client catwalkClient, path string, autoUpdate bool) {
	cs.client = client
	cs.cache = newCache[[]catwalk.Provider](path)
	cs.autoUpdate = autoUpdate
	cs.init.Store(true)
}

func (cs *catwalkSync) Get(ctx context.Context) ([]catwalk.Provider, error) {
	if !cs.init.Load() {
		panic("called Get before Init")
	}
	var throwErr error
	cs.once.Do(func() {
		if !cs.autoUpdate {
			logs.Infof("Using embedded Catwalk provider")
			cs.result = embedded.GetAll()
			return
		}
		cached, etag, cachedErr := cs.cache.Get()
		if len(cached) == 0 || cachedErr != nil {
			// if cached file is empty, default to embedded provider
			cached = embedded.GetAll()
		}
		logs.Infof("Fetching Catwalk provider")
		result, err := cs.client.GetProviders(ctx, etag)
		if errors.Is(err, context.DeadlineExceeded) {
			logs.Warnf("Catwalk provider not updated in time: %v", err)
			cs.result = cached
			return
		}
		if errors.Is(err, catwalk.ErrNotModified) {
			logs.Warnf("Catwalk provider not modified：%v", err)
			cs.result = cached
			return
		}
		if err != nil {
			// On error, fall back to cached (which defaults to embedded if empty).
			cs.result = cached
			return
		}
		if len(result) == 0 {
			cs.result = cached
			throwErr = errors.New("empty provider list from catwalk")
			return
		}
		cs.result = result
		throwErr = cs.cache.Store(result)
	})
	return cs.result, throwErr
}
