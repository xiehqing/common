package entity

import "github.com/xiehqing/common/pkg/ormx"

type UserTenant struct {
	ormx.BaseModel
	UserID   int64   `json:"userId" gorm:"column:user_id;type:bigint;not null"`
	User     *User   `json:"user" gorm:"foreignKey:UserID"`
	TenantID int64   `json:"tenantId" gorm:"column:tenant_id;type:bigint;not null"`
	Tenant   *Tenant `json:"tenant" gorm:"foreignKey:TenantID"`
	RoleID   int64   `json:"roleId" gorm:"column:role_id;type:bigint;not null"`
	Role     *Role   `json:"role" gorm:"foreignKey:RoleID"`
}

func (ut *UserTenant) TableName() string {
	return "user_tenant"
}

type TenantRoleCode string

const (
	TenantRoleOfAdmin TenantRoleCode = "tenant_admin"
	TenantRoleOfOwner TenantRoleCode = "tenant_owner"
)

// IsAdmin 是否为租户管理员,所有者也是管理员
func (ut *UserTenant) IsAdmin() bool {
	return ut.Role.Code == string(TenantRoleOfAdmin) || ut.Role.Code == string(TenantRoleOfOwner)
}

// IsOwner 是否为租户所有者
func (ut *UserTenant) IsOwner() bool {
	return ut.Role.Code == string(TenantRoleOfOwner)
}
