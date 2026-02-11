package entity

import "github.com/xiehqing/common/pkg/ormx"

type Role struct {
	ormx.BaseModel
	Name     string `json:"name" gorm:"column:name;type:varchar(255);not null"`
	Code     string `json:"code" gorm:"column:code;type:varchar(255);not null"`
	IsAdmin  int    `json:"isAdmin" gorm:"column:is_admin;type:tinyint;not null"`
	Comment  string `json:"comment" gorm:"column:comment;type:varchar(255);"`
	TenantID int64  `json:"tenantId" gorm:"column:tenant_id;type:bigint;not null"` // 租户角色，若为0则为系统角色
}

func (r *Role) TableName() string {
	return "role"
}

// Admin 是否为管理员
func (r *Role) Admin() bool {
	return r.IsAdmin == 1
}
