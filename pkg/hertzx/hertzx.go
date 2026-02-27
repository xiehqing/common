package hertzx

import (
	"fmt"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/hertz-contrib/cors"
	"github.com/xiehqing/common/pkg/hertzx/middleware"
	"github.com/xiehqing/common/pkg/resp"
	"net/http"
	"time"
)

type WebConfig struct {
	Host                string `json:"host" yaml:"host"` // 当前主机地址，默认 0.0.0.0
	Port                int    `json:"port" yaml:"port"`
	MaxRequestBodySize  int    `json:"maxRequestBodySize" yaml:"max-request-body-size"`
	ReadTimeout         int    `json:"readTimeout" yaml:"read-timeout" mapstructure:"read-timeout"`    // 读取超时时间，默认 10s
	WriteTimeout        int    `json:"writeTimeout" yaml:"write-timeout" mapstructure:"write-timeout"` // 写入超时时间，默认 10s
	IdleTimeout         int    `json:"idleTimeout" yaml:"idle-timeout" mapstructure:"idle-timeout"`    // 空闲超时时间，默认 120s
	ShutdownTimeout     int    `json:"shutdownTimeout" yaml:"shutdown-timeout" mapstructure:"shutdown-timeout"`
	EnableAPIForService bool   `json:"enableAPIForService" yaml:"enable-api-for-service" mapstructure:"enable-api-for-service"`
}

func (cfg *WebConfig) Prepare() {
	if cfg.Host == "" {
		cfg.Host = "0.0.0.0"
	}
	if cfg.Port == 0 {
		cfg.Port = 8080
	}
	if cfg.MaxRequestBodySize == 0 {
		cfg.MaxRequestBodySize = 1024 * 1024 * 200
	}
	if cfg.ReadTimeout == 0 {
		cfg.ReadTimeout = 3 * 60 * 1000
	}
	if cfg.IdleTimeout == 0 {
		cfg.IdleTimeout = 3 * 60 * 1000
	}
	if cfg.WriteTimeout == 0 {
		cfg.WriteTimeout = 24 * 60 * 60 * 1000
	}
	if cfg.ShutdownTimeout == 0 {
		cfg.ShutdownTimeout = 10 * 1000
	}
}

func WebEngine(cfg WebConfig) *server.Hertz {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	opts := []config.Option{
		server.WithHostPorts(addr),
		server.WithMaxRequestBodySize(cfg.MaxRequestBodySize),
		server.WithReadTimeout(time.Duration(cfg.ReadTimeout) * time.Millisecond),
		server.WithWriteTimeout(time.Duration(cfg.WriteTimeout) * time.Millisecond),
		server.WithIdleTimeout(time.Duration(cfg.IdleTimeout) * time.Millisecond),
		server.WithExitWaitTime(time.Duration(cfg.ShutdownTimeout) * time.Millisecond),
	}
	hertz := server.Default(opts...)

	corsCfg := cors.DefaultConfig()
	corsCfg.AllowAllOrigins = true
	corsCfg.AllowHeaders = []string{"*"}

	hertz.Use(middleware.SetLogIdMW())
	hertz.Use(cors.New(corsCfg))
	hertz.Use(middleware.AccessLogMW())
	return hertz
}

func StartWebServer(hertz *server.Hertz) func() {
	hertz.Spin()
	return func() {}
}

func Bad(c *app.RequestContext, message string) {
	c.AbortWithStatusJSON(http.StatusBadRequest, resp.Response{
		Code:    resp.BadRequest,
		Message: message,
	})
}

// Badf 返回错误信息
func Badf(c *app.RequestContext, format string, args ...interface{}) {
	c.AbortWithStatusJSON(http.StatusBadRequest, resp.Response{
		Code:    resp.BadRequest,
		Message: fmt.Sprintf(format, args...),
	})
}

// OK 返回成功信息
func OK(c *app.RequestContext, data interface{}) {
	c.JSON(http.StatusOK, data)
}

func Data(c *app.RequestContext, data interface{}) {
	c.JSON(http.StatusOK, resp.Success(data))
}

func Msg(c *app.RequestContext, data string) {
	c.JSON(http.StatusOK, resp.Message(data))
}

func Msgf(c *app.RequestContext, format string, args ...interface{}) {
	c.JSON(http.StatusOK, resp.Message(fmt.Sprintf(format, args)))
}

func Error(c *app.RequestContext, message string) {
	c.JSON(http.StatusOK, resp.Error(resp.Failed, message))
}

func Errorf(c *app.RequestContext, format string, args ...interface{}) {
	c.JSON(http.StatusOK, resp.Error(resp.Failed, fmt.Sprintf(format, args)))
}

func Abort(c *app.RequestContext, code int, message string) {
	c.AbortWithStatusJSON(code, resp.Message(message))
}

func Abortf(c *app.RequestContext, code int, format string, args ...interface{}) {
	c.AbortWithStatusJSON(code, resp.Message(fmt.Sprintf(format, args)))
}

func Unauthorized(c *app.RequestContext, message string) {
	Abort(c, 401, message)
}

func Unauthorizedf(c *app.RequestContext, format string, args ...interface{}) {
	Abortf(c, 401, format, args...)
}
