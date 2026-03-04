package entity

import "github.com/xiehqing/common/pkg/ormx"

type UserTenant struct {
	ormx.BaseModel
	UserID   int64   `json:"userId" gorm:"column:user_id;type:bigint;not null"`
	User     *User   `json:"user" gorm:"foreignKey:UserID"`
	TenantID int64   `json:"tenantId" gorm:"column:tenant_id;type:bigint;not null"`
	Tenant   *Tenant `json:"tenant" gorm:"foreignKey:TenantID"`
}

func (ut *UserTenant) TableName() string {
	return "user_tenant"
}
