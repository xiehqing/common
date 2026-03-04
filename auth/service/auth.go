package service

import (
	"github.com/pkg/errors"
	"github.com/xiehqing/common/auth/entity"
	"github.com/xiehqing/common/pkg/logs"
	"github.com/xiehqing/common/pkg/ormx"
	"gorm.io/gorm"
)

type BaseService struct {
}

func NewAuthService() *BaseService {
	return &BaseService{}
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
	userIds := ormx.GetObjIDs(users)
	var permissionMap = make(map[int64]*UserPermission)
	permissions, err := bs.GetMultiUserPermissions(db, userIds)
	if err != nil {
		logs.Errorf("获取用户权限错误:%s", err.Error())
	} else {
		for _, p := range permissions {
			permissionMap[p.UserID] = p
		}
	}
	for i := 0; i < len(users); i++ {
		u := &User{
			ID:        users[i].ID,
			Username:  users[i].Username,
			NickName:  users[i].NickName,
			Email:     users[i].Email,
			Phone:     users[i].Phone,
			Password:  users[i].Password,
			Avatar:    users[i].Avatar,
			Gender:    users[i].Gender,
			Birthday:  users[i].Birthday,
			Signature: users[i].Signature,
		}
		if p, ok := permissionMap[users[i].ID]; ok {
			u.Permission = p
		}
		userRecords = append(userRecords, u)
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

// GetUserPermissions 获取用户权限
func (bs *BaseService) GetUserPermissions(db *gorm.DB, userID int64) (*UserPermission, error) {
	permissions, err := bs.GetMultiUserPermissions(db, []int64{userID})
	if err != nil {
		return nil, err
	}
	if len(permissions) == 0 {
		return nil, errors.Errorf("未发现有效用户.")
	}
	return permissions[0], nil
}

// GetMultiUserPermissions 获取多个用户权限
func (bs *BaseService) GetMultiUserPermissions(db *gorm.DB, userIDs []int64) ([]*UserPermission, error) {
	var users []*entity.User
	var userRoles []*entity.UserRole
	var userTenants []*entity.UserTenant
	var roleOperations []*entity.RoleOperation
	err := db.Where("id in (?)", userIDs).Find(&users).Error
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, errors.Errorf("未发现有效用户.")
	}
	err = db.Where("user_id in (?)", userIDs).Preload("Role").Find(&userRoles).Error
	if err != nil {
		return nil, errors.WithMessagef(err, "获取用户角色列表错误.")
	}
	err = db.Where("user_id in (?)", userIDs).Preload("Tenant").Find(&userTenants).Error
	if err != nil {
		return nil, errors.WithMessagef(err, "获取用户租户列表错误.")
	}
	err = db.Preload("Operation").Find(&roleOperations).Error
	if err != nil {
		return nil, errors.WithMessagef(err, "获取角色权限列表错误.")
	}
	var userPermissions = make([]*UserPermission, 0)
	for _, user := range users {
		userPermissions = append(userPermissions, getUserPermission(user.ID, userRoles, userTenants, roleOperations))
	}
	return userPermissions, nil
}

// getUserPermission 获取用户权限
func getUserPermission(userId int64,
	userRoles []*entity.UserRole,
	userTenants []*entity.UserTenant,
	roleOperations []*entity.RoleOperation,
) *UserPermission {
	// 角色权限
	var roleOptsMap = make(map[int64][]*entity.RoleOperation)
	for _, ro := range roleOperations {
		roleOptsMap[ro.RoleID] = append(roleOptsMap[ro.RoleID], ro)
	}
	var filterUserRoles []*entity.UserRole
	var filterUserTenants []*entity.UserTenant
	for _, r := range userRoles {
		if r.UserID == userId {
			filterUserRoles = append(filterUserRoles, r)
		}
	}
	for _, t := range userTenants {
		if t.UserID == userId {
			filterUserTenants = append(filterUserTenants, t)
		}
	}
	// 租户角色
	var tenantPermissionMap = make(map[int64]map[int64]*RolePermission) // key : tenantId  value: map[roleId]rolePermission
	var systemPermissionMap = make(map[int64]*RolePermission)           // key : roleId  value: rolePermission
	for _, ur := range filterUserRoles {
		if ur.Role != nil && ur.Role.TenantID == 0 {
			var ops []string
			if ros, ok := roleOptsMap[ur.Role.ID]; ok {
				for _, ro := range ros {
					ops = append(ops, ro.Operation.Name)
				}
			}
			// 系统级权限
			if _, ok := systemPermissionMap[ur.Role.ID]; !ok {
				systemPermissionMap[ur.Role.ID] = &RolePermission{
					RoleID: ur.Role.ID,
					Role: &Role{
						ID:      ur.Role.ID,
						Name:    ur.Role.Name,
						Code:    ur.Role.Code,
						IsAdmin: ur.Role.Admin(),
						Comment: ur.Role.Comment,
					},
					Operations: ops,
				}
			}
		} else if ur.Role != nil && ur.Role.TenantID != 0 {
			var ops []string
			if ros, ok := roleOptsMap[ur.Role.ID]; ok {
				for _, ro := range ros {
					ops = append(ops, ro.Operation.Name)
				}
			}
			// 组合权限是否存在此租户的信息
			if ute, ok := tenantPermissionMap[ur.Role.TenantID]; ok {
				// 租户级权限，租户下是否存在此角色
				if _, exist := ute[ur.Role.ID]; !exist {
					tenantPermissionMap[ur.Role.TenantID][ur.Role.ID] = &RolePermission{
						RoleID: ur.Role.ID,
						Role: &Role{
							ID:       ur.Role.ID,
							Name:     ur.Role.Name,
							Code:     ur.Role.Code,
							IsAdmin:  ur.Role.Admin(),
							Comment:  ur.Role.Comment,
							TenantID: ur.Role.TenantID,
						},
						Operations: ops,
					}
				}
			} else {
				tenantPermissionMap[ur.Role.TenantID] = make(map[int64]*RolePermission)
				tenantPermissionMap[ur.Role.TenantID][ur.Role.ID] = &RolePermission{
					RoleID: ur.Role.ID,
					Role: &Role{
						ID:       ur.Role.ID,
						Name:     ur.Role.Name,
						Code:     ur.Role.Code,
						IsAdmin:  ur.Role.Admin(),
						Comment:  ur.Role.Comment,
						TenantID: ur.Role.TenantID,
					},
					Operations: ops,
				}
			}
		}
	}

	var systemPermissions = make([]*RolePermission, 0)
	for _, sp := range systemPermissionMap {
		systemPermissions = append(systemPermissions, sp)
	}
	var tenantPermissions = make([]*TenantPermission, 0)
	for _, ut := range filterUserTenants {
		var roles = make([]*RolePermission, 0)
		if ute, ok := tenantPermissionMap[ut.TenantID]; ok {
			for _, rp := range ute {
				roles = append(roles, rp)
			}
		}
		tenantPermissions = append(tenantPermissions, &TenantPermission{
			TenantID: ut.TenantID,
			Tenant: &Tenant{
				ID:      ut.Tenant.ID,
				Name:    ut.Tenant.Name,
				Code:    ut.Tenant.Code,
				DBName:  ut.Tenant.DBName,
				Comment: ut.Tenant.Comment,
			},
			Roles: roles,
		})
	}
	return &UserPermission{
		UserID:            userId,
		SystemPermissions: systemPermissions,
		TenantPermissions: tenantPermissions,
	}
}

// GetUserTenantsByUserIDs 获取用户租户列表
func (bs *BaseService) GetUserTenantsByUserIDs(db *gorm.DB, userIDs []int64) ([]*entity.UserTenant, error) {
	var uts []*entity.UserTenant
	err := db.Where("user_id in ?", userIDs).
		Preload("Tenant").
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

// GetWxUserByOpenId 获取微信用户信息
func (bs *BaseService) GetWxUserByOpenId(db *gorm.DB, openId string, unionId string) (*entity.WxUser, error) {
	var wxUser *entity.WxUser
	err := db.Where("open_id = ?", openId).First(&wxUser).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, errors.WithMessagef(err, "获取微信用户信息失败")
	}
	return wxUser, nil
}

// GetUserProfile 获取用户基础信息
func (bs *BaseService) GetUserProfile(db *gorm.DB, userID int64) (*BaseUserInfo, error) {
	var user *entity.User
	err := db.Where("id = ?", userID).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, errors.WithMessagef(err, "获取用户信息失败")
	}
	return convertToBaseUserInfo(user), nil
}

// UpdateUserProfile 更新用户基础信息
func (bs *BaseService) UpdateUserProfile(db *gorm.DB, userId int64, req UpdateUserProfileRequest) error {
	var user *entity.User
	err := db.Where("id = ?", userId).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("用户不存在")
		}
		return errors.WithMessagef(err, "获取用户信息失败")
	}
	if req.NickName != nil {
		user.NickName = *req.NickName
	}
	if req.Email != nil {
		user.Email = *req.Email
	}
	if req.Phone != nil {
		user.Phone = *req.Phone
	}
	if req.Avatar != nil {
		user.Avatar = *req.Avatar
	}
	if req.Gender != nil {
		user.Gender = *req.Gender
	}
	if req.Birthday != nil {
		user.Birthday = *req.Birthday
	}
	if req.Signature != nil {
		user.Signature = *req.Signature
	}
	err = ormx.Upsert(db, user)
	if err != nil {
		return errors.WithMessagef(err, "更新用户信息失败")
	}
	return nil
}
