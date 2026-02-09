package ormx

import (
	"fmt"
	"github.com/hatcher/common/pkg/logs"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"sync"
)

type TenantDBManager struct {
	sync.RWMutex
	Master    *gorm.DB
	TenantDBs map[string]*gorm.DB
	Config    *DBConfig
	Handlers  []func(db *gorm.DB)
}

func NewTenantDBManager(master *gorm.DB, config *DBConfig, handlers ...func(db *gorm.DB)) *TenantDBManager {
	return &TenantDBManager{
		Master:    master,
		TenantDBs: make(map[string]*gorm.DB),
		Config:    config,
		Handlers:  handlers,
	}
}

// GetTenantDB 获取租户数据库连接, 如果不存在则创建, 可执行函数
func (dm *TenantDBManager) GetTenantDB(tenantDBName string, handlers ...func(db *gorm.DB)) (*gorm.DB, error) {
	dm.RLock()
	if db, ok := dm.TenantDBs[tenantDBName]; ok {
		for _, handler := range handlers {
			handler(db)
		}
		dm.RUnlock()
		return db, nil
	}
	dm.RUnlock()
	// 创建新的租户数据库连接
	dm.RLock()
	defer dm.RUnlock()
	// 双重检查
	if db, ok := dm.TenantDBs[tenantDBName]; ok {
		return db, nil
	}
	if err := dm.CreateTenantDatabase(tenantDBName); err != nil {
		return nil, err
	}
	tenantDBCfg := dm.copyDBConfigForTenant(tenantDBName)
	tenantDBClient, err := NewDBClient(*tenantDBCfg)
	if err != nil {
		return nil, errors.WithMessagef(err, "初始化租户数据库连接失败")
	}
	dm.TenantDBs[tenantDBName] = tenantDBClient
	for _, handler := range handlers {
		handler(tenantDBClient)
	}
	return tenantDBClient, nil
}

// CreateTenantDatabase 创建租户数据库
func (dm *TenantDBManager) CreateTenantDatabase(dbName string) error {
	var charset = "utf8mb4"
	if dm.Config.Charset != "" {
		charset = dm.Config.Charset
	}
	sql := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s CHARACTER SET %s COLLATE %s_general_ci", dbName, charset, charset)
	if err := dm.Master.Exec(sql).Error; err != nil {
		return errors.WithMessagef(err, "创建租户数据库失败")
	}
	logs.Infof("创建租户数据库成功：%s", dbName)
	return nil
}

func (dm *TenantDBManager) copyDBConfigForTenant(dbName string) *DBConfig {
	return dm.Config.CopyWithDbName(dbName)
}

func (dm *TenantDBManager) Close() error {
	dm.Lock()
	for dbName, dbClient := range dm.TenantDBs {
		if dbClient != nil {
			if sqlDB, err := dbClient.DB(); err == nil {
				ce := sqlDB.Close()
				if ce != nil {
					logs.Errorf("关闭租户数据库连接失败：%s, 错误：%s", dbName, ce.Error())
					return ce
				} else {
					delete(dm.TenantDBs, dbName)
					logs.Infof("关闭租户数据库连接成功：%s", dbName)
				}
			}
		}
	}
	dm.Unlock()
	return nil
}
