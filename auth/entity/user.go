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
	LastActiveTime *time.Time `json:"lastActiveTime" gorm:"column:last_active_time;type:dateTime;"`
}

func (u *User) TableName() string {
	return "users"
}
