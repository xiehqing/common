package jwtx

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/xiehaiqing/common/pkg/jwtx/store"
	"time"
)

type Jwt struct {
	SigningKey     string
	AccessExpired  int64
	RefreshExpired int64
	Store          store.Store
}

// NewJwtService 创建jwt服务
func NewJwtService(cfg Config) (*Jwt, error) {
	ts, err := store.NewJwtStore(cfg.TokenStore)
	if err != nil {
		return nil, err
	}
	return &Jwt{
		SigningKey:     cfg.SigningKey,
		AccessExpired:  cfg.AccessExpired,
		RefreshExpired: cfg.RefreshExpired,
		Store:          ts,
	}, nil
}

type Config struct {
	SigningKey     string       `json:"signingKey" yaml:"signing-key" mapstructure:"signing-key"`
	AccessExpired  int64        `json:"accessExpired" yaml:"access-expired" mapstructure:"access-expired"`
	RefreshExpired int64        `json:"refreshExpired" yaml:"refresh-expired" mapstructure:"refresh-expired"`
	TokenStore     store.Config `json:"tokenStore" yaml:"token-store" mapstructure:"token-store"`
}

type TokenDetails struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	AccessUuid   string `json:"access_uuid"`
	RefreshUuid  string `json:"refresh_uuid"`
	AtExpires    int64  `json:"at_expires"`
	RtExpires    int64  `json:"rt_expires"`
}

type AccessDetails struct {
	AccessUuid   string
	UserIdentity string
}

func wrapJwtKey(prefix, key string) string {
	return fmt.Sprintf("%s:%s", prefix, key)
}

// CreateTokens 创建token
func (j *Jwt) CreateTokens(userIdentity string) (*TokenDetails, error) {
	td := &TokenDetails{}
	signingKey := j.SigningKey
	td.AtExpires = time.Now().Add(time.Minute * time.Duration(j.AccessExpired)).Unix()
	td.AccessUuid = uuid.NewString()

	td.RtExpires = time.Now().Add(time.Minute * time.Duration(j.RefreshExpired)).Unix()
	td.RefreshUuid = td.AccessUuid + "++" + userIdentity

	var err error
	// Creating Access Token
	atClaims := jwt.MapClaims{}
	atClaims["authorized"] = true
	atClaims["access_uuid"] = td.AccessUuid
	atClaims["user_identity"] = userIdentity
	atClaims["exp"] = td.AtExpires
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	td.AccessToken, err = at.SignedString([]byte(signingKey))
	if err != nil {
		return nil, err
	}

	// Creating Refresh Token
	rtClaims := jwt.MapClaims{}
	rtClaims["refresh_uuid"] = td.RefreshUuid
	rtClaims["user_identity"] = userIdentity
	rtClaims["exp"] = td.RtExpires
	jrt := jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaims)
	td.RefreshToken, err = jrt.SignedString([]byte(signingKey))
	if err != nil {
		return nil, err
	}

	return td, nil
}

func (j *Jwt) VerifyToken(signingKey, tokenString string) (*jwt.Token, error) {
	if tokenString == "" {
		return nil, fmt.Errorf("bearer token not found")
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected jwt signing method: %v", token.Header["alg"])
		}
		return []byte(signingKey), nil
	})
	if err != nil {
		return nil, err
	}
	return token, nil
}

// createAuth 创建auth
func (j *Jwt) CreateAuth(ctx context.Context, jwtTokenPrefix, userIdentity string, td *TokenDetails) error {
	at := time.Unix(td.AtExpires, 0)
	rte := time.Unix(td.RtExpires, 0)
	now := time.Now()
	err := j.Store.SaveAccessToken(ctx, wrapJwtKey(jwtTokenPrefix, td.AccessUuid), userIdentity, at.Sub(now))
	if err != nil {
		return err
	}
	err = j.Store.SaveRefreshToken(ctx, wrapJwtKey(jwtTokenPrefix, td.RefreshUuid), userIdentity, rte.Sub(now))
	if err != nil {
		return err
	}
	return nil
}

func (j *Jwt) ExtractToken(tokenStr string) (*AccessDetails, error) {
	token, err := j.VerifyToken(j.SigningKey, tokenStr)
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		accessUuid, exists := claims["access_uuid"].(string)
		if !exists {
			return nil, fmt.Errorf("failed to parse access_uuid from jwt")
		}
		return &AccessDetails{
			AccessUuid:   accessUuid,
			UserIdentity: claims["user_identity"].(string),
		}, nil
	}

	return nil, err
}

// FetchAuth 获取auth
//func (j *Jwt) FetchAuth(ctxparam context.Context, jwtTokenPrefix, givenUuid string) (string, error) {
//	return j.TokenStore.Get(ctxparam, wrapJwtKey(jwtTokenPrefix, givenUuid))
//}

// DeleteAuth 删除auth
//func (j *Jwt) DeleteAuth(ctxparam context.Context, jwtTokenPrefix, givenUuid string) error {
//	return j.TokenStore.Delete(ctxparam, wrapJwtKey(jwtTokenPrefix, givenUuid))
//}

// DeleteTokens 删除token
func (j *Jwt) DeleteTokens(ctx context.Context, jwtTokenPrefix string, authD *AccessDetails) error {
	// get the refresh uuid
	refreshUuid := authD.AccessUuid + "++" + authD.UserIdentity
	// delete access token
	err := j.Store.DeleteAccessToken(ctx, wrapJwtKey(jwtTokenPrefix, authD.AccessUuid))
	if err != nil {
		return err
	}
	// delete refresh token
	err = j.Store.DeleteRefreshToken(ctx, wrapJwtKey(jwtTokenPrefix, refreshUuid))
	if err != nil {
		return err
	}
	return nil
}

// RenewTokenIfNeeded 如果token快要过期则续签
// renewThreshold: 续签阈值（分钟），当token剩余时间少于此值时触发续签
// renewDuration: 续签时长（分钟），续签后token的有效期
// 返回：是否进行了续签
func (j *Jwt) RenewTokenIfNeeded(ctx context.Context, jwtTokenPrefix, tokenStr string, renewThreshold, renewDuration int64) (bool, error) {
	// 解析token
	token, err := j.VerifyToken(j.SigningKey, tokenStr)
	if err != nil {
		return false, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return false, fmt.Errorf("invalid token claims")
	}

	// 获取过期时间
	exp, ok := claims["exp"].(float64)
	if !ok {
		return false, fmt.Errorf("failed to parse exp from token")
	}

	expireTime := time.Unix(int64(exp), 0)
	now := time.Now()
	remainingTime := expireTime.Sub(now)

	// 检查是否需要续签（剩余时间少于阈值）
	if remainingTime.Minutes() > float64(renewThreshold) {
		// 不需要续签
		return false, nil
	}

	// 需要续签，提取必要信息
	accessUuid, ok := claims["access_uuid"].(string)
	if !ok {
		return false, fmt.Errorf("failed to parse access_uuid from token")
	}

	userIdentity, ok := claims["user_identity"].(string)
	if !ok {
		return false, fmt.Errorf("failed to parse user_identity from token")
	}

	// 延长 access token 有效期
	newExpiration := time.Duration(renewDuration) * time.Minute
	err = j.Store.SaveAccessToken(ctx, wrapJwtKey(jwtTokenPrefix, accessUuid), userIdentity, newExpiration)
	if err != nil {
		return false, fmt.Errorf("failed to renew access token: %w", err)
	}

	// 同时延长 refresh token 有效期
	refreshUuid := accessUuid + "++" + userIdentity
	err = j.Store.SaveRefreshToken(ctx, wrapJwtKey(jwtTokenPrefix, refreshUuid), userIdentity, newExpiration)
	if err != nil {
		return false, fmt.Errorf("failed to renew refresh token: %w", err)
	}

	return true, nil
}
