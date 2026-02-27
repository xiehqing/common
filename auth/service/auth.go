package service

import (
	"github.com/pkg/errors"
	"github.com/xiehqing/common/auth/entity"
	"github.com/xiehqing/common/models"
	"github.com/xiehqing/common/pkg/ormx"
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

// GetUserByID 根据ID获取用户
func (bs *BaseService) GetUserByID(db *gorm.DB, id int64) (*User, error) {
	if id == 0 {
		return nil, nil
	}
	users, err := bs.GetUsers(db, "id = ?", id)
	if err != nil {
		return nil, errors.WithMessagef(err, "获取用户信息失败,id:%d", id)
	}
	if len(users) == 0 {
		return nil, nil
	}
	return users[0], nil
}

// GetUserByUsername 根据用户名获取用户
func (bs *BaseService) GetUserByUsername(db *gorm.DB, username string) (*User, error) {
	if username == "" {
		return nil, nil
	}
	users, err := bs.GetUsers(db, "username = ?", username)
	if err != nil {
		return nil, errors.WithMessagef(err, "获取用户信息失败,username:%s", username)
	}
	if len(users) == 0 {
		return nil, nil
	}
	return users[0], nil
}

// GetUserTenantsByUserIDs 获取用户租户列表
func (bs *BaseService) GetUserTenantsByUserIDs(db *gorm.DB, userIDs []int64) ([]*entity.UserTenant, error) {
	var uts []*entity.UserTenant
	err := db.Where("user_id in ?", userIDs).
		Preload("Tenant").
		Preload("Role").
		Preload("User").
		Find(&uts).
		Error
	if err != nil {
		return nil, err
	}
	return uts, nil
}

// GetUserTenantsByUserID 获取用户租户列表
func (bs *BaseService) GetUserTenantsByUserID(db *gorm.DB, userID int64) ([]*entity.UserTenant, error) {
	return bs.GetUserTenantsByUserIDs(db, []int64{userID})
}

// GetUserTenantsByTenantID 获取租户用户列表
func (bs *BaseService) GetUserTenantsByTenantID(db *gorm.DB, tenantID int64) ([]*entity.UserTenant, error) {
	var uts []*entity.UserTenant
	err := db.Where("tenant_id = ?", tenantID).
		Preload("Tenant").
		Preload("Role").
		Preload("User").
		Find(&uts).
		Error
	if err != nil {
		return nil, err
	}
	return uts, nil
}

// GetTenantByID 获取租户信息
func (bs *BaseService) GetTenantByID(db *gorm.DB, id int64) (*entity.Tenant, error) {
	var tenant *entity.Tenant
	err := db.Where("id = ?", id).First(&tenant).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, errors.WithMessagef(err, "获取租户信息失败,id:%d", id)
	}
	return tenant, nil
}

// GetTenantByCode 获取租户信息
func (bs *BaseService) GetTenantByCode(db *gorm.DB, code string) (*entity.Tenant, error) {
	var tenant *entity.Tenant
	err := db.Where("code = ?", code).First(&tenant).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, errors.WithMessagef(err, "获取租户信息失败,code:%s", code)
	}
	return tenant, nil
}

// GetTenantByCodes 获取租户信息
func (bs *BaseService) GetTenantByCodes(db *gorm.DB, codes []string) ([]*entity.Tenant, error) {
	var tenants []*entity.Tenant
	err := db.Where("code in ?", codes).Find(&tenants).Error
	if err != nil {
		return nil, errors.WithMessagef(err, "获取租户信息失败")
	}
	return tenants, nil
}

// GetAllTenants 获取所有租户信息
func (bs *BaseService) GetAllTenants(db *gorm.DB) ([]*entity.Tenant, error) {
	var tenants []*entity.Tenant
	err := db.Find(&tenants).Error
	if err != nil {
		return nil, errors.WithMessagef(err, "获取租户信息失败")
	}
	return tenants, nil
}

// GetConfigValue 获取系统配置
func (bs *BaseService) GetConfigValue(db *gorm.DB, key string) (string, error) {
	var lst []string
	err := db.Model(&entity.SystemConfigs{}).Where("config_key = ?", key).Pluck("config_value", &lst).Error
	if err != nil {
		return "", errors.WithMessagef(err, "获取系统配置:`%s`错误", key)
	}
	if len(lst) == 0 {
		return "", nil
	}
	return lst[0], nil
}

// GetSystemRoles 获取系统角色列表
func (bs *BaseService) GetSystemRoles(db *gorm.DB) ([]*entity.Role, error) {
	var roles []*entity.Role
	err := db.Where("tenant_id = 0").Find(&roles).Error
	if err != nil {
		return nil, errors.WithMessagef(err, "获取系统角色列表错误.")
	}
	return roles, nil
}

// GetTenantRoles 获取租户角色列表
func (bs *BaseService) GetTenantRoles(db *gorm.DB, tenantID int64) ([]*entity.Role, error) {
	var roles []*entity.Role
	err := db.Where("tenant_id = ?", tenantID).Find(&roles).Error
	if err != nil {
		return nil, errors.WithMessagef(err, "获取租户角色列表错误.")
	}
	return roles, nil
}

// CheckRoleOperation 检查角色是否具有操作权限
func (bs *BaseService) CheckRoleOperation(db *gorm.DB, roleIDs []int64, operation string) (bool, error) {
	if len(roleIDs) == 0 {
		return false, nil
	}
	return ormx.Exists(db.Model(&entity.RoleOperation{}).Where("role_id in (?) and operation = ?", roleIDs, operation))
}
