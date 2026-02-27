package service

import (
	"github.com/xiehqing/common/auth/entity"
	"gorm.io/gorm"
)

type User struct {
	ID        int64         `json:"id"`
	Username  string        `json:"username"`
	NickName  string        `json:"nickName"`
	Email     string        `json:"email"`
	Phone     string        `json:"phone"`
	Password  string        `json:"password"`
	Avatar    string        `json:"avatar"`
	Gender    entity.Gender `json:"gender"`
	Birthday  string        `json:"birthday"`
	Signature string        `json:"signature"`
	Roles     []Role        `json:"roles"`
	Tenants   []Tenant      `json:"tenants"`
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

func (u *User) GetRoles() []Role {
	return u.Roles
}

func (u *User) GetTenants() []Tenant {
	return u.Tenants
}

func (u *User) IsAdmin() bool {
	var isAdmin = false
	for _, role := range u.Roles {
		if role.IsAdmin {
			isAdmin = true
			break
		}
	}
	return isAdmin
}

// HasPermission 是否具有权限
func (u *User) HasPermission(db *gorm.DB, permission string) (bool, error) {
	if u.IsAdmin() {
		return true, nil
	}
	var roleIds []int64
	for _, role := range u.GetRoles() {
		roleIds = append(roleIds, role.ID)
	}
	return entity.CheckRoleOperation(db, roleIds, permission)
}

func (u *User) CheckTenant(db *gorm.DB, tenantID int64) (*Tenant, error) {
	tenants := u.GetTenants()
	for _, tenant := range tenants {
		if tenant.ID == tenantID {
			return &tenant, nil
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
