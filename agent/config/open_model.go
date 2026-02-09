package config

import (
	"cmp"
	"encoding/json"
	"fmt"
	"github.com/charmbracelet/catwalk/pkg/catwalk"
	"github.com/hatcher/common/agent/home"
	"github.com/hatcher/common/pkg/logs"
	"os"
	"path/filepath"
	"runtime"
)

const openProviderName = "open_provider"

// OpenProviderData 返回开放模型提供者
func OpenProviderData() string {
	if openProviderData := os.Getenv(EnvOpenProviderData); openProviderData != "" {
		return filepath.Join(openProviderData, fmt.Sprintf("%s.json", openProviderName))
	}
	if runtime.GOOS == "windows" {
		localAppData := cmp.Or(os.Getenv(EnvLocalAppData),
			filepath.Join(os.Getenv(EnvUserProfile), "AppData", "Local"))
		return filepath.Join(localAppData, appName, fmt.Sprintf("%s.json", openProviderName))
	}
	return filepath.Join(home.Dir(), ".local", "share", appName, fmt.Sprintf("%s.json", openProviderName))
}

func OpenProviders(cfg *Config) ([]catwalk.Provider, error) {
	openProviderFile := OpenProviderData()
	logs.Infof("Open Provider provider file: %s", openProviderFile)
	var openProviders []catwalk.Provider
	if openProviderFile != "" {
		bytes, err := os.ReadFile(openProviderFile)
		if err != nil {
			return openProviders, fmt.Errorf("failed to read Open Provider provider file: %w", err)
		}
		if err := json.Unmarshal(bytes, &openProviders); err != nil {
			return openProviders, fmt.Errorf("failed to unmarshal Open Provider provider file: %w", err)
		}
	} else {
		logs.Warnf("no Open Provider provider file found")
	}
	return openProviders, nil
}
