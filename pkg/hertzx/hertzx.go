package hertzx

import (
	"fmt"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/hertz-contrib/cors"
	"github.com/pkg/errors"
	"github.com/xiehqing/common/models"
	"github.com/xiehqing/common/pkg/hertzx/middleware"
	"github.com/xiehqing/common/pkg/resp"
	"github.com/xiehqing/common/pkg/util"
	"net/http"
	"strconv"
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

// ParamInt64 获取参数
func ParamInt64(c *app.RequestContext, paramName string) (int64, error) {
	paramContent := c.Param(paramName)
	if paramContent == "" {
		return 0, fmt.Errorf("参数 %s 不能为空", paramName)
	}
	return strconv.ParseInt(paramContent, 10, 64)
}

// ParamInt 获取参数
func ParamInt(c *app.RequestContext, paramName string) (int, error) {
	paramContent := c.Param(paramName)
	if paramContent == "" {
		return 0, fmt.Errorf("参数 %s 不能为空", paramName)
	}
	return strconv.Atoi(paramContent)
}

// QueryInt64 获取int64参数
func QueryInt64(c *app.RequestContext, paramName string) (int64, error) {
	pv := c.DefaultQuery(paramName, "")
	var v int64
	if pv == "" {
		return v, nil
	}
	return strconv.ParseInt(pv, 10, 64)
}

func DefaultQueryInt64(c *app.RequestContext, paramName string, defaultValue int64) (int64, error) {
	pv := c.DefaultQuery(paramName, "")
	if pv == "" {
		return defaultValue, nil
	}
	return strconv.ParseInt(pv, 10, 64)
}

// QueryInt64Ptr 获取int64参数
func QueryInt64Ptr(c *app.RequestContext, paramName string) (*int64, error) {
	pv := c.DefaultQuery(paramName, "")
	var v *int64
	if pv != "" {
		vv, err := strconv.ParseInt(pv, 10, 64)
		if err != nil {
			return nil, errors.WithMessagef(err, "参数 %s 转换失败", paramName)
		}
		v = &vv
	}
	return v, nil
}

// QueryInt 获取int参数
func QueryInt(c *app.RequestContext, paramName string) (int, error) {
	pv := c.DefaultQuery(paramName, "")
	var v int
	if pv == "" {
		return v, nil
	}
	return strconv.Atoi(pv)
}

// QueryIntPtr 获取int参数
func QueryIntPtr(c *app.RequestContext, paramName string) (*int, error) {
	pv := c.DefaultQuery(paramName, "")
	var v *int
	if pv != "" {
		vv, err := strconv.Atoi(pv)
		if err != nil {
			return nil, errors.WithMessagef(err, "参数 %s 转换失败", paramName)
		}
		v = &vv
	}
	return v, nil
}

// QueryDatePtr 获取date参数
func QueryDatePtr(c *app.RequestContext, paramName string) (*time.Time, error) {
	pv := c.DefaultQuery(paramName, "")
	var v *time.Time
	if pv != "" {
		vv, err := util.ParseTime("2006-01-02 15:04:05", pv)
		if err != nil {
			return nil, errors.WithMessagef(err, "参数 %s 转换失败", paramName)
		}
		v = &vv
	}
	return v, nil
}

// ParsePageable 解析分页参数
func ParsePageable(c *app.RequestContext) (models.Pageable, error) {
	pageNo, err := QueryInt(c, "pageNo")
	pageable := models.Pageable{}
	if err != nil {
		return pageable, errors.WithMessagef(err, "参数 pageNo 不合法")
	}
	pageSize, err := QueryInt(c, "pageSize")
	if err != nil {
		return pageable, errors.WithMessagef(err, "参数 pageSize 不合法")
	}
	sortField := c.DefaultQuery("sortField", "updated_at")
	if sortField == "" {
		sortField = "updated_at"
	}
	sortOrder := c.DefaultQuery("sortOrder", "desc")
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}
	return models.PageRequest(pageNo, pageSize, sortField, sortOrder), nil
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
