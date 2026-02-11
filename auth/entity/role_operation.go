package entity

import "github.com/xiehqing/common/pkg/ormx"

type RoleOperation struct {
	ormx.BaseModel
	RoleID    int64  `json:"roleId" gorm:"column:role_id;type:bigint;not null"`
	Role      *Role  `json:"role" gorm:"foreignKey:RoleID"`
	Operation string `json:"operation" gorm:"column:operation;type:varchar(200);not null"`
}

func (r *RoleOperation) TableName() string {
	return "role_operation"
}
