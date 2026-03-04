package entity

import "github.com/xiehqing/common/pkg/ormx"

type Operation struct {
	ormx.BaseModel
	Name        string `json:"name" gorm:"column:name;type:varchar(50);not null"`
	DisplayName string `json:"displayName" gorm:"column:display_name;type:varchar(50);not null"`
}

func (o *Operation) TableName() string {
	return "operation"
}
