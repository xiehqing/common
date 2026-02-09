package store

import (
	"context"
	"gorm.io/gorm"
	"time"
)

type MysqlStore struct {
	DB *gorm.DB
}

func (m *MysqlStore) SaveAccessToken(ctx context.Context, key string, value string, expiration time.Duration) error {
	return nil
}
func (m *MysqlStore) GetAccessToken(ctx context.Context, key string) (string, error) {
	return "", nil
}
func (m *MysqlStore) DeleteAccessToken(ctx context.Context, key string) error {
	return nil
}
func (m *MysqlStore) SaveRefreshToken(ctx context.Context, key string, value string, expiration time.Duration) error {
	return nil
}

func (m *MysqlStore) GetRefreshToken(ctx context.Context, key string) (string, error) {
	return "", nil
}

func (m *MysqlStore) DeleteRefreshToken(ctx context.Context, key string) error {
	return nil
}

func (m *MysqlStore) GetLoginFailedCount(ctx context.Context, key string) (int64, error) {
	return 0, nil
}

func (m *MysqlStore) IncrLoginFailedCount(ctx context.Context, key string, expiration time.Duration) error {
	return nil
}
