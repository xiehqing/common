package entity

import "github.com/xiehqing/common/pkg/ormx"

type UserRole struct {
	ormx.BaseModel
	UserID int64 `json:"userId" gorm:"column:user_id;type:bigint;not null"`
	User   *User `json:"user" gorm:"foreignKey:UserID"`
	RoleID int64 `json:"roleId" gorm:"column:role_id;type:bigint;not null"`
	Role   *Role `json:"role" gorm:"foreignKey:RoleID"`
}

func (ur *UserRole) TableName() string {
	return "user_role"
}
