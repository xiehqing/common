package service

import (
	"github.com/xiehqing/common/auth/entity"
	"gorm.io/gorm/utils"
)

type User struct {
	ID         int64           `json:"id"`
	Username   string          `json:"username"`
	NickName   string          `json:"nickName"`
	Email      string          `json:"email"`
	Phone      string          `json:"phone"`
	Password   string          `json:"password"`
	Avatar     string          `json:"avatar"`
	Gender     entity.Gender   `json:"gender"`
	Birthday   string          `json:"birthday"`
	Signature  string          `json:"signature"`
	Permission *UserPermission `json:"permission"`
}

type BaseUserInfo struct {
	ID             int64         `json:"id"`
	Username       string        `json:"username"`
	NickName       string        `json:"nickName"`
	Email          string        `json:"email"`
	Phone          string        `json:"phone"`
	Password       string        `json:"-"`
	Avatar         string        `json:"avatar"`
	Gender         entity.Gender `json:"gender"`
	Birthday       string        `json:"birthday"`
	Signature      string        `json:"signature"`
	LastActiveTime string        `json:"lastActiveTime"`
	CreatedAt      string        `json:"createdAt"`
	UpdatedAt      string        `json:"updatedAt"`
}

func (u *User) GetSystemRoles() []*RolePermission {
	var roles = make([]*RolePermission, 0)
	if u.Permission != nil && len(u.Permission.SystemPermissions) > 0 {
		roles = u.Permission.SystemPermissions
	}
	return roles
}

func (u *User) GetTenants() []*TenantPermission {
	var tenants = make([]*TenantPermission, 0)
	if u.Permission != nil && len(u.Permission.TenantPermissions) > 0 {
		tenants = u.Permission.TenantPermissions
	}
	return tenants
}

func (u *User) IsSystemAdmin() bool {
	return u.Permission != nil && u.Permission.IsSystemAdmin()
}

// HasPermission 是否具有权限
func (u *User) HasPermission(tenantID int64, permission string) (bool, error) {
	// tenantID 为0的时候判断系统权限
	if tenantID == 0 {
		if u.IsSystemAdmin() {
			return true, nil
		} else {
			var hasPermission = false
			if u.Permission != nil && len(u.Permission.SystemPermissions) > 0 {
				for _, role := range u.Permission.SystemPermissions {
					if len(role.Operations) > 0 && utils.Contains(role.Operations, permission) {
						hasPermission = true
					}
				}
			}
			return hasPermission, nil
		}
	} else {
		var hasTenantPermission = false
		if u.Permission != nil && len(u.Permission.TenantPermissions) > 0 {
			for _, tenant := range u.Permission.TenantPermissions {
				if tenant.TenantID == tenantID {
					if tenant.IsAdmin() {
						hasTenantPermission = true
					} else {
						for _, role := range tenant.Roles {
							if len(role.Operations) > 0 && utils.Contains(role.Operations, permission) {
								hasTenantPermission = true
							}
						}
					}
				}
			}
		}
		return hasTenantPermission, nil
	}
}

// HasSystemPermission 是否具有系统权限
func (u *User) HasSystemPermission(permission string) (bool, error) {
	return u.HasPermission(0, permission)
}

// CheckTenant 检查用户是否具有租户权限
func (u *User) CheckTenant(tenantID int64) (*Tenant, error) {
	tenants := u.GetTenants()
	for _, tenant := range tenants {
		if tenant.TenantID == tenantID {
			return tenant.Tenant, nil
		}
	}
	return nil, nil
}

type Role struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Code     string `json:"code"`
	IsAdmin  bool   `json:"isAdmin"`
	Comment  string `json:"comment"`
	TenantID int64  `json:"tenantId"`
}

type Tenant struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Code    string `json:"code"`
	DBName  string `json:"dbName"`
	Comment string `json:"comment"`
	IsAdmin bool   `json:"isAdmin"`
}

// 转换为基础用户信息
func convertToBaseUserInfo(user *entity.User) *BaseUserInfo {
	b := &BaseUserInfo{
		ID:        user.ID,
		Username:  user.Username,
		NickName:  user.NickName,
		Avatar:    user.Avatar,
		Phone:     user.Phone,
		Email:     user.Email,
		Signature: user.Signature,
		Gender:    user.Gender,
		Birthday:  user.Birthday,
	}
	if user.LastActiveTime != nil {
		b.LastActiveTime = user.LastActiveTime.Format("2006-01-02 15:04:05")
	}
	if user.CreatedAt != nil {
		b.CreatedAt = user.CreatedAt.Format("2006-01-02 15:04:05")
	}
	if user.UpdatedAt != nil {
		b.UpdatedAt = user.UpdatedAt.Format("2006-01-02 15:04:05")
	}
	return b
}

type UpdateUserProfileRequest struct {
	NickName  *string        `json:"nickName"`
	Email     *string        `json:"email"`
	Phone     *string        `json:"phone"`
	Avatar    *string        `json:"avatar"`
	Gender    *entity.Gender `json:"gender"`
	Birthday  *string        `json:"birthday"`
	Signature *string        `json:"signature"`
}

type RolePermission struct {
	RoleID     int64    `json:"roleId"`
	Role       *Role    `json:"role"`
	Operations []string `json:"operations"`
}

func (rp *RolePermission) IsAdmin() bool {
	return rp.Role != nil && rp.Role.IsAdmin
}

type TenantPermission struct {
	TenantID int64             `json:"tenantId"`
	Tenant   *Tenant           `json:"tenant"`
	Roles    []*RolePermission `json:"roles"`
}

// IsAdmin 是否为租户管理员
func (tp *TenantPermission) IsAdmin() bool {
	var isAdmin = false
	if len(tp.Roles) > 0 {
		for _, r := range tp.Roles {
			if r.IsAdmin() {
				isAdmin = true
				break
			}
		}
	}
	return isAdmin
}

type UserPermission struct {
	UserID            int64               `json:"userId"`
	SystemPermissions []*RolePermission   `json:"systemPermissions"`
	TenantPermissions []*TenantPermission `json:"tenantPermissions"`
}

// IsSystemAdmin 是否为系统管理员
func (up *UserPermission) IsSystemAdmin() bool {
	var isAdmin = false
	if len(up.SystemPermissions) > 0 {
		for _, r := range up.SystemPermissions {
			if r.IsAdmin() {
				isAdmin = true
				break
			}
		}
	}
	return isAdmin
}
