package service

import "gorm.io/gorm"

type User struct {
	ID       int64    `json:"id"`
	Username string   `json:"username"`
	NickName string   `json:"nickName"`
	Email    string   `json:"email"`
	Phone    string   `json:"phone"`
	Password string   `json:"password"`
	Roles    []Role   `json:"roles"`
	Tenants  []Tenant `json:"tenants"`
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
