package config

import (
	"context"
	"github.com/charmbracelet/catwalk/pkg/catwalk"
	"testing"
)

type mockCatwalkClient struct {
	providers []catwalk.Provider
	err       error
	callCount int
}

func (m *mockCatwalkClient) GetProviders(ctx context.Context, etag string) ([]catwalk.Provider, error) {
	m.callCount++
	return m.providers, m.err
}

func TestCatwalkSync_Init(t *testing.T) {
	syncer := &catwalkSync{}
	client := &mockCatwalkClient{}
	path := "/tmp/test.json"
	syncer.Init(client, path, true)
	p := syncer.cache.path
	t.Log(p)
}
