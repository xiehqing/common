package ormx

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"strings"
	"time"
)

// DBConfig 数据库配置
type DBConfig struct {
	Debug              bool   `yaml:"debug" json:"debug" mapstructure:"debug"`
	DbType             string `yaml:"db-type" json:"dbType" mapstructure:"db-type"`
	DSN                string `yaml:"dsn" json:"dsn" mapstructure:"dsn"`
	Host               string `yaml:"host" json:"host" mapstructure:"host"`
	Port               int    `yaml:"port" json:"port" mapstructure:"port"`
	Username           string `yaml:"username" json:"username" mapstructure:"username"`
	Password           string `yaml:"password" json:"password" mapstructure:"password"`
	Database           string `yaml:"database" json:"database" mapstructure:"database"`
	Charset            string `yaml:"charset" json:"charset" mapstructure:"charset"`
	AppendParams       string `yaml:"append-params" json:"appendParams" mapstructure:"append-params"`
	MaxLifetime        int    `yaml:"max-lifetime" json:"maxLifetime" mapstructure:"max-lifetime"`
	MaxOpenConnections int    `yaml:"max-open-connections" json:"maxOpenConnections" mapstructure:"max-open-connections"`
	MaxIdleConnections int    `yaml:"max-idle-connections" json:"maxIdleConnections" mapstructure:"max-idle-connections"`
	TablePrefix        string `yaml:"table-prefix" json:"tablePrefix" mapstructure:"table-prefix"`
}

// GetDSNByDBName 获取指定名称数据库连接字符串
func (c *DBConfig) GetDSNByDBName(dbName string) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s",
		c.Username,
		c.Password,
		c.Host,
		c.Port,
		dbName,
		c.AppendParams,
	)
}

// GetDSN 获取数据库连接字符串
func (c *DBConfig) GetDSN() string {
	if c.DSN != "" {
		return c.DSN
	}
	if c.AppendParams == "" {
		c.AppendParams = "parseTime=True&loc=Local"
	}
	if c.Charset != "" && !strings.Contains(c.AppendParams, "charset") {
		c.AppendParams += fmt.Sprintf("&charset=%s", c.Charset)
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s",
		c.Username,
		c.Password,
		c.Host,
		c.Port,
		c.Database,
		c.AppendParams,
	)
}

// CopyWithDbName 复制配置并指定数据库名称
func (c *DBConfig) CopyWithDbName(dbName string) *DBConfig {
	if c == nil {
		return nil
	}
	return &DBConfig{
		Host:               c.Host,
		Port:               c.Port,
		Username:           c.Username,
		Password:           c.Password,
		Database:           dbName,
		AppendParams:       c.AppendParams,
		MaxLifetime:        c.MaxLifetime,
		MaxIdleConnections: c.MaxIdleConnections,
		MaxOpenConnections: c.MaxOpenConnections,
		TablePrefix:        c.TablePrefix,
		DbType:             c.DbType,
		Debug:              c.Debug,
	}
}

// NewDBClient 创建db客户端
func NewDBClient(c DBConfig) (*gorm.DB, error) {
	var dialect gorm.Dialector
	switch strings.ToLower(c.DbType) {
	case "mysql":
		dialect = mysql.Open(c.GetDSN())
	default:
		return nil, fmt.Errorf("dialector(%s) not supported", c.DbType)
	}
	gormConfig := &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   c.TablePrefix,
			SingularTable: true,
		},
	}
	db, err := gorm.Open(dialect, gormConfig)
	if err != nil {
		return nil, err
	}
	if c.Debug {
		db = db.Debug()
	}
	sqlDb, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDb.SetMaxIdleConns(c.MaxIdleConnections)
	sqlDb.SetMaxOpenConns(c.MaxOpenConnections)
	sqlDb.SetConnMaxLifetime(time.Duration(c.MaxLifetime) * time.Second)
	return db, nil
}

// BaseModel 基础model
type BaseModel struct {
	ID        int64      `json:"id" gorm:"primaryKey;autoIncrement"`
	CreatedAt *time.Time `json:"createdAt" gorm:"type:dateTime;autoCreateTime;not null;comment:'创建时间'"`
	CreatedBy string     `json:"createdBy" gorm:"type:varchar(255);not null;comment:'创建人'"`
	UpdatedAt *time.Time `json:"updatedAt" gorm:"type:dateTime;autoUpdateTime;not null;comment:'更新时间'"`
	UpdatedBy string     `json:"updatedBy" gorm:"type:varchar(255);not null;comment:'更新人'"`
}

// UuidModel 基础model
type UuidModel struct {
	ID        string     `json:"id" gorm:"primaryKey"`
	CreatedAt *time.Time `json:"createdAt" gorm:"type:dateTime;autoCreateTime;not null;comment:'创建时间'"`
	CreatedBy string     `json:"createdBy" gorm:"type:varchar(255);not null;comment:'创建人'"`
	UpdatedAt *time.Time `json:"updatedAt" gorm:"type:dateTime;autoUpdateTime;not null;comment:'更新时间'"`
	UpdatedBy string     `json:"updatedBy" gorm:"type:varchar(255);not null;comment:'更新人'"`
}

// DeleteAbleModel 逻辑删除的model
type DeleteAbleModel struct {
	BaseModel
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// StatusAbleModel 带状态的model
type StatusAbleModel struct {
	BaseModel
	Status int `json:"status" gorm:"type:int(11);not null;comment:'状态'"`
}

type Pagination struct {
	Keyword   string `json:"keyword" form:"keyword"`
	PageNo    int    `json:"pageNo" form:"pageNo"`
	PageSize  int    `json:"pageSize" form:"pageSize"`
	SortField string `json:"sortField" form:"sortField"`
	SortOrder string `json:"sortOrder" form:"sortOrder"`
}
