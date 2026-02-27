package entity

import (
	"github.com/xiehqing/common/pkg/ormx"
)

func (w *WxUser) TableName() string {
	return "wx_users"
}

type WxUser struct {
	ormx.BaseModel
	OpenId  string `json:"openId" gorm:"column:open_id;type:varchar(255);not null"`
	UnionId string `json:"unionId" gorm:"column:union_id;type:varchar(255);not null"`
	UserID  int64  `json:"userId" gorm:"column:user_id;type:bigint;not null"`
}
