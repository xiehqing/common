package config

import (
	"github.com/hatcher/common/agent/provider"
	"sync/atomic"
)

var instance atomic.Pointer[Config]

func Init(workingDir, dataDir string, pvdSvc provider.Service, debug bool) (*Config, error) {
	cfg, err := Load(workingDir, dataDir, debug, pvdSvc)
	if err != nil {
		return nil, err
	}
	instance.Store(cfg)
	return instance.Load(), nil
}

func Get() *Config {
	cfg := instance.Load()
	return cfg
}

// InitAndGet 初始化并获取配置
func InitAndGet(workingDir, dataDir string, pvdSvc provider.Service, debug, skipRequests bool) (*Config, error) {
	cfg, err := Load(workingDir, dataDir, debug, pvdSvc)
	if err != nil {
		return nil, err
	}
	instance.Store(cfg)
	load := instance.Load()
	if cfg.Permissions == nil {
		cfg.Permissions = &Permissions{SkipRequests: skipRequests}
	} else {
		cfg.Permissions.SkipRequests = skipRequests
	}
	return load, nil
}
