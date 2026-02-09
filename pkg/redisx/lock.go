package redisx

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"golang.org/x/net/context"
	"time"
)

// DistributedLock 分布式锁
type DistributedLock struct {
	client        Redis
	key           string
	value         string
	expiration    time.Duration
	watchdog      chan struct{} // 用于停止自动续期
	autoRenew     bool          // 是否自动续期
	renewInterval time.Duration // 续期间隔
	maxRetryCount int           // 最大重试次数
	retryInterval time.Duration // 重试间隔
}

// LockOptions 锁配置选项
type LockOptions struct {
	Key           string        // 锁的key
	Value         string        // 锁的value（用于标识锁持有者，为空则自动生成UUID）
	Expiration    time.Duration // 锁过期时间，默认30秒
	AutoRenew     bool          // 是否自动续期，默认false
	RenewInterval time.Duration // 续期间隔，默认为过期时间的1/3
	MaxRetryCount int           // 获取锁的最大重试次数，默认3次
	RetryInterval time.Duration // 重试间隔，默认100ms
}

// NewDistributedLock 创建分布式锁（兼容旧接口）
func NewDistributedLock(client Redis, key string, value string, expiration *time.Duration) *DistributedLock {
	exp := 30 * time.Second
	if expiration != nil {
		exp = *expiration
	}

	if value == "" {
		value = uuid.New().String()
	}

	return &DistributedLock{
		client:        client,
		key:           key,
		value:         value,
		expiration:    exp,
		autoRenew:     false,
		renewInterval: exp / 3,
		maxRetryCount: 3,
		retryInterval: 100 * time.Millisecond,
	}
}

// NewDistributedLockWithOptions 使用配置选项创建分布式锁
func NewDistributedLockWithOptions(client Redis, opts LockOptions) *DistributedLock {
	// 设置默认值
	if opts.Key == "" {
		opts.Key = "distributed_lock:" + uuid.New().String()
	}
	if opts.Value == "" {
		opts.Value = uuid.New().String()
	}
	if opts.Expiration == 0 {
		opts.Expiration = 30 * time.Second
	}
	if opts.RenewInterval == 0 {
		opts.RenewInterval = opts.Expiration / 3
	}
	if opts.MaxRetryCount == 0 {
		opts.MaxRetryCount = 3
	}
	if opts.RetryInterval == 0 {
		opts.RetryInterval = 100 * time.Millisecond
	}

	return &DistributedLock{
		client:        client,
		key:           opts.Key,
		value:         opts.Value,
		expiration:    opts.Expiration,
		autoRenew:     opts.AutoRenew,
		renewInterval: opts.RenewInterval,
		maxRetryCount: opts.MaxRetryCount,
		retryInterval: opts.RetryInterval,
	}
}

// TryLock 尝试获取锁（非阻塞）
func (l *DistributedLock) TryLock(ctx context.Context) (bool, error) {
	result, err := l.client.SetNX(ctx, l.key, l.value, l.expiration).Result()
	if err != nil {
		return false, errors.WithMessagef(err, "获取锁失败")
	}

	// 如果获取成功且需要自动续期，启动watchdog
	if result && l.autoRenew {
		l.startWatchdog(ctx)
	}

	return result, nil
}

// Lock 阻塞式获取锁，带重试机制
func (l *DistributedLock) Lock(ctx context.Context) error {
	for i := 0; i < l.maxRetryCount; i++ {
		acquired, err := l.TryLock(ctx)
		if err != nil {
			return err
		}

		if acquired {
			return nil
		}

		// 如果是最后一次重试，直接返回失败
		if i == l.maxRetryCount-1 {
			break
		}

		// 等待后重试
		select {
		case <-ctx.Done():
			return errors.New("获取锁被取消")
		case <-time.After(l.retryInterval):
			continue
		}
	}

	return errors.Errorf("获取锁失败，已重试 %d 次", l.maxRetryCount)
}

// LockWithTimeout 带超时的阻塞式获取锁
func (l *DistributedLock) LockWithTimeout(ctx context.Context, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(l.retryInterval)
	defer ticker.Stop()

	for {
		acquired, err := l.TryLock(ctx)
		if err != nil {
			return err
		}

		if acquired {
			return nil
		}

		select {
		case <-ctx.Done():
			return errors.New("获取锁超时或被取消")
		case <-ticker.C:
			continue
		}
	}
}

// Unlock 释放锁（使用Lua脚本保证原子性，只能释放自己持有的锁）
func (l *DistributedLock) Unlock(ctx context.Context) error {
	// 停止自动续期
	if l.watchdog != nil {
		close(l.watchdog)
		l.watchdog = nil
	}

	// Lua脚本：只有value匹配时才删除key
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`

	result, err := l.client.Eval(ctx, script, []string{l.key}, l.value).Result()
	if err != nil {
		return errors.WithMessagef(err, "释放锁失败")
	}

	if result == int64(0) {
		return errors.New("释放锁失败：锁不存在或已被其他持有者占用")
	}

	return nil
}

// Refresh 刷新锁的过期时间
func (l *DistributedLock) Refresh(ctx context.Context) error {
	// Lua脚本：只有value匹配时才更新过期时间
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("pexpire", KEYS[1], ARGV[2])
		else
			return 0
		end
	`

	result, err := l.client.Eval(ctx, script, []string{l.key}, l.value, l.expiration.Milliseconds()).Result()
	if err != nil {
		return errors.WithMessagef(err, "刷新锁失败")
	}

	if result == int64(0) {
		return errors.New("刷新锁失败：锁不存在或已被其他持有者占用")
	}

	return nil
}

// IsLocked 检查锁是否被持有
func (l *DistributedLock) IsLocked(ctx context.Context) (bool, error) {
	val, err := l.client.Get(ctx, l.key).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, errors.WithMessagef(err, "检查锁状态失败")
	}
	return val != "", nil
}

// IsLockedByMe 检查锁是否被当前实例持有
func (l *DistributedLock) IsLockedByMe(ctx context.Context) (bool, error) {
	val, err := l.client.Get(ctx, l.key).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, errors.WithMessagef(err, "检查锁状态失败")
	}
	return val == l.value, nil
}

// GetTTL 获取锁的剩余生存时间
func (l *DistributedLock) GetTTL(ctx context.Context) (time.Duration, error) {
	ttl, err := l.client.TTL(ctx, l.key).Result()
	if err != nil {
		return 0, errors.WithMessagef(err, "获取锁TTL失败")
	}
	return ttl, nil
}

// startWatchdog 启动看门狗，自动续期
func (l *DistributedLock) startWatchdog(ctx context.Context) {
	l.watchdog = make(chan struct{})

	go func() {
		ticker := time.NewTicker(l.renewInterval)
		defer ticker.Stop()

		for {
			select {
			case <-l.watchdog:
				return
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := l.Refresh(ctx); err != nil {
					// 续期失败，说明锁可能已经被释放或被其他持有者占用
					return
				}
			}
		}
	}()
}

// ExecuteWithLock 在锁保护下执行函数
func (l *DistributedLock) ExecuteWithLock(ctx context.Context, fn func() error) error {
	// 获取锁
	if err := l.Lock(ctx); err != nil {
		return errors.WithMessage(err, "获取锁失败")
	}

	// 确保函数执行完后释放锁
	defer func() {
		if err := l.Unlock(ctx); err != nil {
			// 记录错误但不影响函数执行结果
			fmt.Printf("释放锁失败: %v\n", err)
		}
	}()

	// 执行业务逻辑
	return fn()
}

// ExecuteWithLockTimeout 在锁保护下执行函数（带超时）
func (l *DistributedLock) ExecuteWithLockTimeout(ctx context.Context, timeout time.Duration, fn func() error) error {
	// 获取锁
	if err := l.LockWithTimeout(ctx, timeout); err != nil {
		return errors.WithMessage(err, "获取锁失败")
	}

	// 确保函数执行完后释放锁
	defer func() {
		if err := l.Unlock(ctx); err != nil {
			fmt.Printf("释放锁失败: %v\n", err)
		}
	}()

	// 执行业务逻辑
	return fn()
}
