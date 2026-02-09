package store

import (
	"context"
	"fmt"
	"github.com/hatcher/common/pkg/ormx"
	"github.com/hatcher/common/pkg/redisx"
	"github.com/hatcher/common/pkg/util"
	"time"
)

type Store interface {
	SaveAccessToken(ctx context.Context, key string, value string, expiration time.Duration) error
	GetAccessToken(ctx context.Context, key string) (string, error)
	DeleteAccessToken(ctx context.Context, key string) error
	SaveRefreshToken(ctx context.Context, key string, value string, expiration time.Duration) error
	GetRefreshToken(ctx context.Context, key string) (string, error)
	DeleteRefreshToken(ctx context.Context, key string) error
	GetLoginFailedCount(ctx context.Context, key string) (int64, error)
	IncrLoginFailedCount(ctx context.Context, key string, expiration time.Duration) error
}

type Config struct {
	Type   string                 `json:"type" yaml:"type" mapstructure:"type"`
	Option map[string]interface{} `json:"option" yaml:"option" mapstructure:"option"`
}

func NewJwtStore(cfg Config) (Store, error) {
	switch cfg.Type {
	case "db":
		dbConfig, err := util.Convert[ormx.DBConfig](cfg.Option)
		if err != nil {
			return nil, err
		}
		db, err := ormx.NewDBClient(*dbConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to initial mysql token store client: %s", err)
		}
		return &MysqlStore{DB: db}, nil
	case "redis":
		redisConfig, err := util.Convert[redisx.RedisConfig](cfg.Option)
		if err != nil {
			return nil, err
		}
		redis, err := redisx.NewRedis(*redisConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to initial redis token store client: %s", err)
		}
		return &RedisStore{RedisCli: redis}, nil
	default:
		return nil, fmt.Errorf("failed to initial token store client: %s", cfg.Type)
	}
}
