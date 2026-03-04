package entity

import "github.com/xiehqing/common/pkg/ormx"

func (ual *UserActivityLog) TableName() string {
	return "user_activity_log"
}

type UserActivityLog struct {
	ormx.BaseModel
	UserID      int64  `json:"userId" gorm:"column:user_id;type:bigint;not null"`
	Action      string `json:"action" gorm:"column:action;type:varchar(500);not null"`
	Description string `json:"description" gorm:"column:description;type:varchar(2500);not null"`
	IP          string `json:"ip" gorm:"column:ip;type:varchar(255);not null"`
	UserAgent   string `json:"userAgent" gorm:"column:user_agent;type:varchar(255);not null"`
}
