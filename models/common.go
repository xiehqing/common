package models

import (
	"github.com/hatcher/common/pkg/logs"
	"github.com/hatcher/common/pkg/util"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"reflect"
	"strings"
)

func Count(tx *gorm.DB) (int64, error) {
	var cnt int64
	err := tx.Count(&cnt).Error
	return cnt, err
}

func Exists(tx *gorm.DB) (bool, error) {
	num, err := Count(tx)
	return num > 0, err
}

func Insert(tx *gorm.DB, obj interface{}) error {
	return tx.Create(obj).Error
}

func CreateInBatches(tx *gorm.DB, obj interface{}) error {
	return CreateInBatchesWithBatchSize(tx, obj, 2000)
}

func CreateInBatchesWithBatchSize(tx *gorm.DB, obj interface{}, batchSize int) error {
	return tx.CreateInBatches(obj, batchSize).Error
}

func Update(tx *gorm.DB, obj interface{}) error {
	return tx.Save(obj).Error
}

func Delete(tx *gorm.DB, obj interface{}) error {
	return tx.Delete(obj).Error
}

// GetByCondition 通用的数据库查询方法
func GetByCondition[T interface{}](db *gorm.DB, where string, args ...interface{}) ([]T, error) {
	var lst []T
	err := db.Where(where, args...).Find(&lst).Error
	if err != nil {
		return nil, err
	}
	return lst, nil
}

func ModelStatistics[T interface{}](db *gorm.DB) (*Statistics, error) {
	session := db.Model(new(T)).Select("count(*) as total", "max(updated_at) as last_updated")
	var stats []*Statistics
	err := session.Find(&stats).Error
	if err != nil {
		return nil, err
	}
	return stats[0], nil
}

type Statistics struct {
	Total       int64 `gorm:"total"`
	LastUpdated int64 `gorm:"last_updated"`
}

type Pageable struct {
	PageNo   int       `json:"pageNo"`
	PageSize int       `json:"pageSize"`
	Sortable *Sortable `json:"sortable"`
}

func PageRequest(pageNo, pageSize int, sortField, sortOrder string) Pageable {
	if pageNo <= 0 {
		pageNo = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	var sa = &Sortable{}
	if sortOrder == "" {
		sa.SortOrder = "asc"
	} else {
		sa.SortOrder = strings.ToLower(sortOrder)
	}

	if sortField == "" {
		sa.SortField = "id"
	} else {
		sa.SortField = sortField
	}
	return Pageable{
		PageNo:   pageNo,
		PageSize: pageSize,
		Sortable: sa,
	}
}

func (pa *Pageable) Offset() int {
	if pa.PageNo <= 0 {
		pa.PageNo = 1
	}
	if pa.PageSize <= 0 {
		pa.PageSize = 10
	}
	return (pa.PageNo - 1) * pa.PageSize
}

type Sortable struct {
	SortField string `json:"sortField"`
	SortOrder string `json:"sortOrder"`
}

func (sa *Sortable) Sort() string {
	var sortOrder, sortField string
	if sa.SortOrder == "" {
		sortOrder = "asc"
	} else {
		sortOrder = strings.ToLower(sa.SortOrder)
	}

	if sa.SortField == "" {
		sortField = "id"
	} else {
		sortField = sa.SortField
	}

	return sortField + " " + sortOrder
}

// PageQuery 分页查询
func PageQuery[T interface{}](tx *gorm.DB, pageable *Pageable, where string, args ...interface{}) ([]T, int64, error) {
	var lst []T
	var total int64
	query := tx.Model(new(T))
	if where != "" {
		query = query.Where(where, args...)
	}
	e := query.Count(&total).Error
	if e != nil {
		logs.Errorf("Page 统计失败: %v", e)
		return nil, 0, e
	}
	if pageable.Sortable != nil {
		query = query.Order(pageable.Sortable.Sort())
	}
	err := query.Offset(pageable.Offset()).
		Limit(pageable.PageSize).
		Find(&lst).
		Error
	if err != nil {
		logs.Errorf("Page 查询失败: %v", err)
		return nil, 0, err
	}
	return lst, total, nil
}

// Upsert 更新或写入
func Upsert(db *gorm.DB, obj interface{}) error {
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return errors.Errorf("当前对象不支持更新或写入.")
	}
	statusField := val.FieldByName("ID")
	if !statusField.IsValid() || !statusField.CanSet() {
		statusField = val.FieldByName("Id")
		if !statusField.IsValid() || !statusField.CanSet() {
			statusField = val.FieldByName("id")
			if !statusField.IsValid() || !statusField.CanSet() {
				return errors.Errorf("当前对象无状态信息.")
			}
		}
	}
	if statusField.IsZero() {
		return db.Create(obj).Error
	} else {
		return db.Save(obj).Error
	}
}

// GetObjIDs 获取对象的 ID 列表
func GetObjIDs(objs interface{}) []int64 {
	ids, has := GetObjIDWithField(objs, "ID")
	if has {
		return ids
	}
	return []int64{}
}

// GetObjIDWithField 根据字段获取ids
func GetObjIDWithField(obj interface{}, fieldName string) ([]int64, bool) {
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	var ids []int64
	hasField := false
	if val.Kind() == reflect.Struct {
		field := val.FieldByName(fieldName)
		if field.IsValid() && field.CanInterface() {
			hasField = true
			id, err := util.ToInt64(field.Interface())
			if err != nil {
				logs.Errorf("转换id字段失败:%v, 错误：%v", field.Interface(), err)
				return nil, false
			}
			ids = append(ids, id)
		}
		return ids, hasField
	}

	if val.Kind() == reflect.Slice {
		if val.Len() == 0 {
			return nil, hasField
		}
		for i := 0; i < val.Len(); i++ {
			elem := val.Index(i)
			if elem.Kind() == reflect.Ptr {
				elem = elem.Elem()
			}
			if elem.Kind() != reflect.Struct {
				continue
			}
			field := elem.FieldByName(fieldName)
			if field.IsValid() && field.CanInterface() {
				hasField = true
				id, err := util.ToInt64(field.Interface())
				if err != nil {
					logs.Errorf("转换id字段失败:%v, 错误：%v", field.Interface(), err)
					continue
				}
				ids = append(ids, id)
			}
		}
		return ids, hasField
	}
	return nil, false
}

// UpdateStatus 更新对象的状态
func UpdateStatus(db *gorm.DB, obj interface{}, status int) error {
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return errors.Errorf("当前对象不支持更新状态.")
	}
	// 查找Status字段
	statusField := val.FieldByName("Status")
	if !statusField.IsValid() || !statusField.CanSet() {
		statusField = val.FieldByName("status")
		if !statusField.IsValid() || !statusField.CanSet() {
			return errors.Errorf("当前对象无状态信息.")
		}
	}
	// 设置状态值
	if statusField.Kind() == reflect.Int || statusField.Kind() == reflect.Int64 {
		statusField.SetInt(int64(status))
	} else {
		return errors.Errorf("状态类型错误.")
	}
	// 保存到数据库
	return db.Save(obj).Error
}

type Counter struct {
	ID    int64 `json:"id" gorm:"column:id"`
	Total int64 `json:"total" gorm:"column:total"`
}
