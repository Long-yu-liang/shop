package initialize

import (
	"fmt"
	"shop_servs/user_srv/global"

	"github.com/spf13/viper"
)

func GetEnvInfo(env string) bool {
	viper.AutomaticEnv()
	return viper.GetBool(env)
}

func InitConfig() {
	debug := GetEnvInfo("SHOP_DEBUG")
	configFilePrefix := "config"
	configFileName := fmt.Sprintf("user_srv/%s-pro.yaml", configFilePrefix)
	if debug {
		configFileName = fmt.Sprintf("user_srv/%s-debug.yaml", configFilePrefix)
	}

	v := viper.New()
	v.SetConfigFile(configFileName)

	fmt.Printf("尝试加载配置文件: %s\n", configFileName) // 调试输出

	if err := v.ReadInConfig(); err != nil {
		panic(fmt.Sprintf("配置文件不存在或格式错误: %v", err))
	}

	if err := v.Unmarshal(&global.ServerConfig); err != nil {
		panic(err)
	}
}
