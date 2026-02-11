package entity

import "github.com/xiehqing/common/pkg/ormx"

type SystemConfigs struct {
	ormx.BaseModel
	Key     string `json:"configKey" gorm:"column:config_key;type:varchar(255);not null"`
	Value   string `json:"configValue" gorm:"column:config_value;type:text;not null"`
	Comment string `json:"comment" gorm:"column:comment;type:varchar(2000);"`
}

func (s *SystemConfigs) TableName() string {
	return "system_configs"
}
