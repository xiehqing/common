package config

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hatcher/common/agent/env"
	"github.com/hatcher/common/agent/fsext"
	"github.com/hatcher/common/agent/provider"
	"github.com/hatcher/common/pkg/logs"
	"github.com/hatcher/common/pkg/util"
	"github.com/qjebbs/go-jsons"
	"os"
	"slices"
)

func Load(workingDir, dataDir string, debug bool, pvdSvc provider.Service) (*Config, error) {
	configPaths := lookupConfigs(workingDir)
	cfg, err := loadFromConfigPaths(configPaths)
	if err != nil {
		return nil, fmt.Errorf("failed to load config from paths %v: %w", configPaths, err)
	}
	cfg.dataConfigDir = GlobalConfigData()
	cfg.setDefaults(workingDir, dataDir)
	if debug {
		cfg.Options.Debug = debug
	}

	providers, err := Providers(cfg)
	if err != nil {
		return nil, err
	}

	openProviders, err := OpenProviders(cfg)
	if err == nil {
		logs.Debugf("Found %d open provider:%s", len(openProviders), util.ToJsonIgnoreError(openProviders))
		providers = append(providers, openProviders...)
	}
	// 从数据库里获取
	dbProviders, err := pvdSvc.List(context.Background())
	if err != nil {
		logs.Errorf("failed to get db providers: %v", err)
	} else {
		logs.Debugf("Found %d db provider:%s", len(dbProviders), util.ToJsonIgnoreError(dbProviders))
		providers = append(providers, dbProviders...)
	}

	cfg.knownProviders = providers
	env := env.New()
	valueResolver := NewShellVariableResolver(env)
	cfg.resolver = valueResolver
	if err := cfg.configureProviders(env, valueResolver, cfg.knownProviders); err != nil {
		return nil, fmt.Errorf("failed to configure provider: %w", err)
	}
	if !cfg.IsConfigured() {
		logs.Warnf("no provider configured")
		return cfg, nil
	}
	if err := cfg.configureSelectedModels(cfg.knownProviders); err != nil {
		return nil, fmt.Errorf("failed to configure selected models: %w", err)
	}
	cfg.SetupAgents()
	return cfg, nil
}

// lookupConfigs 从当前工作目录开始，递归向上搜索配置文件直到文件系统根目录
func lookupConfigs(cwd string) []string {
	configPaths := []string{
		GlobalConfig(),
		GlobalConfigData(),
	}
	logs.Infof("configPaths:%s", util.ToJsonIgnoreError(configPaths))
	configNames := []string{appName + ".json", "." + appName + ".json"}
	foundConfigs, err := fsext.Lookup(cwd, configNames...)
	if err != nil {
		logs.Errorf("failed to lookup config files: %v", err)
		return configPaths
	}
	slices.Reverse(foundConfigs)
	return append(configPaths, foundConfigs...)
}

// loadFromConfigPaths 从配置路径中加载配置
func loadFromConfigPaths(configPaths []string) (*Config, error) {
	var configs [][]byte
	for _, path := range configPaths {
		data, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("failed to open config file %s: %w", path, err)
		}
		if len(data) == 0 {
			continue
		}
		configs = append(configs, data)
	}
	return loadFromBytes(configs)
}

// loadFromBytes 从字节数组中加载配置
func loadFromBytes(configs [][]byte) (*Config, error) {
	if len(configs) == 0 {
		return &Config{}, nil
	}

	data, err := jsons.Merge(configs)
	if err != nil {
		return nil, err
	}
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}
