package redisx

import (
	"context"
	"errors"
	"fmt"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/xiehaiqing/common/pkg/logs"
	"github.com/xiehaiqing/common/pkg/tlsx"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	ImhLoginTokenPrefix     = "imh-authtoken"
	ImhLoginUserErrorPrefix = "imh-loginfailedcount"
)

const UserLoginERRORCountPrefix = "/userlogin/errorcount/"

type RedisConfig struct {
	Address  string `json:"address" mapstructure:"address" yaml:"address"`
	Username string `json:"username" mapstructure:"username" yaml:"username"`
	Password string `json:"password" mapstructure:"password" yaml:"password"`
	DB       int    `json:"db" mapstructure:"db" yaml:"db"`
	tlsx.ClientConfig
	RedisType        string `json:"redisType" mapstructure:"redis-type" yaml:"redis-type"`
	MasterName       string `json:"masterName" mapstructure:"master-name" yaml:"master-name"`
	SentinelUsername string `json:"sentinelUsername" mapstructure:"sentinel-username" yaml:"sentinel-username"`
	SentinelPassword string `json:"sentinelPassword" mapstructure:"sentinel-password" yaml:"sentinel-password"`
}

type Redis redis.Cmdable

func NewRedis(cfg RedisConfig) (Redis, error) {
	var redisClient Redis

	switch cfg.RedisType {
	case "standalone", "":
		redisOptions := &redis.Options{
			Addr:     cfg.Address,
			Username: cfg.Username,
			Password: cfg.Password,
			DB:       cfg.DB,
		}

		if cfg.UseTLS {
			tlsConfig, err := cfg.TLSConfig()
			if err != nil {
				logs.Errorf("failed to initial redisx tls config: %v", err)
				os.Exit(1)
			}
			redisOptions.TLSConfig = tlsConfig
		}

		redisClient = redis.NewClient(redisOptions)

	case "cluster":
		redisOptions := &redis.ClusterOptions{
			Addrs:    strings.Split(cfg.Address, ","),
			Username: cfg.Username,
			Password: cfg.Password,
		}

		if cfg.UseTLS {
			tlsConfig, err := cfg.TLSConfig()
			if err != nil {
				logs.Errorf("failed to initial redisx tls config: %v", err)
				os.Exit(1)
			}
			redisOptions.TLSConfig = tlsConfig
		}

		redisClient = redis.NewClusterClient(redisOptions)

	case "sentinel":
		redisOptions := &redis.FailoverOptions{
			MasterName:       cfg.MasterName,
			SentinelAddrs:    strings.Split(cfg.Address, ","),
			Username:         cfg.Username,
			Password:         cfg.Password,
			DB:               cfg.DB,
			SentinelUsername: cfg.SentinelUsername,
			SentinelPassword: cfg.SentinelPassword,
		}

		if cfg.UseTLS {
			tlsConfig, err := cfg.TLSConfig()
			if err != nil {
				logs.Errorf("failed to initial redisx tls config: %v", err)
				os.Exit(1)
			}
			redisOptions.TLSConfig = tlsConfig
		}

		redisClient = redis.NewFailoverClient(redisOptions)

	case "miniredis":
		s, err := miniredis.Run()
		if err != nil {
			logs.Errorf("failed to initial miniredis: %v", err)
			os.Exit(1)
		}
		redisClient = redis.NewClient(&redis.Options{
			Addr: s.Addr(),
		})

	default:
		logs.Infof("failed to initial redisx , redisx type is illegal: %s", cfg.RedisType)
		os.Exit(1)
	}

	err := redisClient.Ping(context.Background()).Err()
	if err != nil {
		logs.Errorf("failed to ping redisx: %v", err)
		os.Exit(1)
	}
	return redisClient, nil
}

func MGet(ctx context.Context, r Redis, keys []string) [][]byte {
	var vals [][]byte
	pipe := r.Pipeline()
	for _, key := range keys {
		pipe.Get(ctx, key)
	}
	cmds, _ := pipe.Exec(ctx)

	for i, key := range keys {
		cmd := cmds[i]
		if errors.Is(cmd.Err(), redis.Nil) {
			continue
		}

		if cmd.Err() != nil {
			logs.Errorf("failed to get key: %s, err: %s", key, cmd.Err())
			continue
		}
		val := []byte(cmd.(*redis.StringCmd).Val())
		vals = append(vals, val)
	}

	return vals
}

func MSet(ctx context.Context, r Redis, m map[string]interface{}) error {
	pipe := r.Pipeline()
	for k, v := range m {
		pipe.Set(ctx, k, v, 0)
	}
	_, err := pipe.Exec(ctx)
	return err
}

// ReachCount 判断key的值是否大于等于count
func ReachCount(ctx context.Context, r Redis, key string, count int64) (bool, error) {
	value, err := r.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	c, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return false, err
	}
	return c >= count, nil
}

// IncrCount 设置key的值count + 1
func IncrCount(ctx context.Context, r Redis, key string, seconds int64) {
	duration := time.Duration(seconds) * time.Second
	value, err := r.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		r.Set(ctx, key, "1", duration)
		return
	}
	if err != nil {
		logs.Warnf("failed to get redis value. key:%s, error:%s", key, err)
		r.Set(ctx, key, "1", duration)
		return
	}
	count, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		logs.Warnf("failed to parse int64. key:%s, error:%s", key, err)
		r.Set(ctx, key, "1", duration)
		return
	}
	count++
	r.Set(ctx, key, fmt.Sprintf("%d", count), duration)
}
