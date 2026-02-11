package service

import (
	"github.com/pkg/errors"
	"github.com/xiehqing/common/auth/entity"
	"github.com/xiehqing/common/models"
	"gorm.io/gorm"
)

type BaseService struct {
}

// GetUsers 获取用户列表
func (bs *BaseService) GetUsers(db *gorm.DB, where string, args ...interface{}) ([]*User, error) {
	var users []*entity.User
	err := db.Where(where, args...).Find(&users).Error
	if err != nil {
		return nil, errors.WithMessagef(err, "获取用户列表错误.")
	}
	var userRecords []*User
	if len(users) == 0 {
		return userRecords, nil
	}
	userIds := models.GetObjIDs(users)
	var userTenants []*entity.UserTenant
	err = db.Where("user_id in (?)", userIds).
		Preload("Tenant").
		Preload("Role").
		Preload("User").
		Find(&userTenants).Error
	if err != nil {
		return nil, errors.WithMessagef(err, "获取用户租户列表错误.")
	}
	var userTenantMap = make(map[int64][]*entity.UserTenant)
	for _, ut := range userTenants {
		userTenantMap[ut.UserID] = append(userTenantMap[ut.UserID], ut)
	}
	for i := 0; i < len(users); i++ {
		if utes, ok := userTenantMap[users[i].ID]; ok {
			var tenants []Tenant
			var roles []Role
			for _, ut := range utes {
				tenants = append(tenants, Tenant{
					ID:      ut.Tenant.ID,
					Name:    ut.Tenant.Name,
					Code:    ut.Tenant.Code,
					DBName:  ut.Tenant.DBName,
					Comment: ut.Tenant.Comment,
					IsAdmin: ut.IsAdmin(),
				})
				roles = append(roles, Role{
					ID:       ut.Role.ID,
					Name:     ut.Role.Name,
					Code:     ut.Role.Code,
					IsAdmin:  ut.Role.Admin(),
					Comment:  ut.Role.Comment,
					TenantID: ut.Tenant.ID,
				})
			}
			userRecords = append(userRecords, &User{
				ID:       users[i].ID,
				Username: users[i].Username,
				NickName: users[i].NickName,
				Email:    users[i].Email,
				Phone:    users[i].Phone,
				Password: users[i].Password,
				Tenants:  tenants,
				Roles:    roles,
			})
		}
	}
	return userRecords, nil
}
