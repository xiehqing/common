package hertzx

import (
	"fmt"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/pkg/errors"
	"github.com/xiehqing/common/models"
	"github.com/xiehqing/common/pkg/util"
	"strconv"
	"time"
)

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
