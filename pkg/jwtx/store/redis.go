package store

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"github.com/toolkits/pkg/logger"
	"github.com/xiehaiqing/common/pkg/logs"
	"github.com/xiehaiqing/common/pkg/redisx"
	"strconv"
	"time"
)

type RedisStore struct {
	RedisCli redisx.Redis
}

func (r *RedisStore) SaveAccessToken(ctx context.Context, key string, value string, expiration time.Duration) error {
	return r.RedisCli.Set(ctx, key, value, expiration).Err()
}
func (r *RedisStore) GetAccessToken(ctx context.Context, key string) (string, error) {
	return r.RedisCli.Get(ctx, key).Result()
}
func (r *RedisStore) DeleteAccessToken(ctx context.Context, key string) error {
	return r.RedisCli.Del(ctx, key).Err()
}
func (r *RedisStore) SaveRefreshToken(ctx context.Context, key string, value string, expiration time.Duration) error {
	return r.RedisCli.Set(ctx, key, value, expiration).Err()
}

func (r *RedisStore) GetRefreshToken(ctx context.Context, key string) (string, error) {
	return r.RedisCli.Get(ctx, key).Result()
}

func (r *RedisStore) DeleteRefreshToken(ctx context.Context, key string) error {
	return r.RedisCli.Del(ctx, key).Err()
}

func (r *RedisStore) GetLoginFailedCount(ctx context.Context, key string) (int64, error) {
	redisx.ReachCount(ctx, r.RedisCli, key, 5)
	value, err := r.RedisCli.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, nil
		}
		return 0, err
	}
	val, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, nil
	}
	return val, nil
}

func (r *RedisStore) IncrLoginFailedCount(ctx context.Context, key string, expiration time.Duration) error {
	value, err := r.RedisCli.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			r.RedisCli.Set(ctx, key, "1", expiration)
			return nil
		}
		logs.Warnf("failed to get redis value. key:%s, error:%s", key, err)
		r.RedisCli.Set(ctx, key, "1", expiration)
		return nil
	}
	count, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		logger.Warningf("failed to parse int64. key:%s, error:%s", key, err)
		r.RedisCli.Set(ctx, key, "1", expiration)
		return nil
	}
	count++
	r.RedisCli.Set(ctx, key, fmt.Sprintf("%d", count), expiration)
	return nil
}
