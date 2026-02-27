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

type SystemConfigKey string

const (
	LoginFailCount       SystemConfigKey = "login_fail_count"         // 登录失败次数,格式 `300 5`
	PwdAesSalt           SystemConfigKey = "pwd_aes_salt"             // 密码加密盐, 若不为空，则认为开启aes认证，为空则认为关闭
	TenantDBNamePrefix   SystemConfigKey = "tenant_db_name_prefix"    // 租户数据库名称前缀
	LoginTokenPrefix     SystemConfigKey = "login_token_store_prefix" // 登录token存储的前缀
	LoginUserErrorPrefix SystemConfigKey = "login_user_error_prefix"
)
