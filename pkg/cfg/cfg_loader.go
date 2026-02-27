package cfg

import (
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

// LoadConfig 加载配置文件
func LoadConfig(configDir, configFile, configSuffix string, ptr interface{}) error {
	viper.SetConfigName(configFile)
	viper.AddConfigPath(configDir)
	viper.SetConfigType(configSuffix)
	err := viper.ReadInConfig()
	if err != nil {
		return errors.WithMessagef(err, "读取配置文件失败，配置文件：%s, 配置目录：%s, 配置文件类型：%s", configFile, configDir, configSuffix)
	}
	err = viper.Unmarshal(ptr)
	if err != nil {
		return errors.WithMessagef(err, "解析配置文件失败，配置文件：%s, 配置目录：%s, 配置文件类型：%s", configFile, configDir, configSuffix)
	}
	return nil
}
