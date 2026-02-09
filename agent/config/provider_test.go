package config

import (
	"sync"
	"testing"
)

func resetProviderState() {
	providerOnce = sync.Once{}
	providerList = nil
	providerErr = nil
	catwalkSyncer = &catwalkSync{}
	//hyperSyncer = &hyperSync{}
}

func TestProviders_Integration_AutoUpdateDisabled(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	// Use a test-specific instance to avoid global state interference.
	testCatwalkSyncer := &catwalkSync{}

	originalCatwalSyncer := catwalkSyncer
	defer func() {
		catwalkSyncer = originalCatwalSyncer
	}()

	catwalkSyncer = testCatwalkSyncer

	resetProviderState()
	defer resetProviderState()

	cfg := &Config{
		Options: &Options{
			DisableProviderAutoUpdate: true,
		},
	}

	providers, err := Providers(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	t.Logf("provider: %v", providers)
}
