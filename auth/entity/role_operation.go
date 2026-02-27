package entity

import (
	"github.com/xiehqing/common/pkg/ormx"
	"gorm.io/gorm"
)

type RoleOperation struct {
	ormx.BaseModel
	RoleID    int64  `json:"roleId" gorm:"column:role_id;type:bigint;not null"`
	Role      *Role  `json:"role" gorm:"foreignKey:RoleID"`
	Operation string `json:"operation" gorm:"column:operation;type:varchar(200);not null"`
}

func (r *RoleOperation) TableName() string {
	return "role_operation"
}

// CheckRoleOperation 检查角色是否具有操作权限
func CheckRoleOperation(db *gorm.DB, roleIDs []int64, operation string) (bool, error) {
	if len(roleIDs) == 0 {
		return false, nil
	}
	return ormx.Exists(db.Model(&RoleOperation{}).Where("role_id in (?) and operation = ?", roleIDs, operation))
}
