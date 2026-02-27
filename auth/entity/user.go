package entity

import (
	"github.com/xiehqing/common/pkg/ormx"
	"time"
)

type User struct {
	ormx.DeleteAbleModel
	Username       string     `json:"username" gorm:"column:username;type:varchar(255);not null"`
	NickName       string     `json:"nickName" gorm:"column:nickName;type:varchar(255);"`
	Password       string     `json:"password" gorm:"column:password;type:varchar(255);not null"`
	Phone          string     `json:"phone" gorm:"column:phone;type:varchar(255);"`
	Email          string     `json:"email" gorm:"column:email;type:varchar(255);"`
	Avatar         string     `json:"avatar" gorm:"column:avatar;type:varchar(3000);"`
	Gender         Gender     `json:"gender" gorm:"column:gender;type:varchar(10);"`
	Birthday       string     `json:"birthday" gorm:"column:birthday;type:varchar(25);"`
	Signature      string     `json:"signature" gorm:"column:signature;type:varchar(5000);"`
	LastActiveTime *time.Time `json:"lastActiveTime" gorm:"column:last_active_time;type:dateTime;"`
}

func (u *User) TableName() string {
	return "users"
}

type Gender string

const (
	Male   Gender = "male"
	Female Gender = "female"
	Secret Gender = "secret"
)
