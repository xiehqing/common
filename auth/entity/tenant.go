package entity

import "github.com/xiehqing/common/pkg/ormx"

type Tenant struct {
	ormx.DeleteAbleModel
	Name    string `json:"name" gorm:"column:name;type:varchar(30);not null"`
	Code    string `json:"code" gorm:"column:code;type:varchar(50);not null"`
	DBName  string `json:"dbName" gorm:"column:dbName;type:varchar(100);not null"`
	Comment string `json:"comment" gorm:"column:comment;type:varchar(2000);"`
}

func (t *Tenant) TableName() string {
	return "tenant"
}
