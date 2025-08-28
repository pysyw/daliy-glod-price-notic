package cfg

import (
	"fmt"

	"github.com/spf13/viper"
)

func init() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.AddConfigPath(".")
	viper.AddConfigPath("../.")
	viper.AddConfigPath("../../.")
	viper.AddConfigPath("../../../.")
	viper.AddConfigPath("../../../../../.")
	viper.AddConfigPath("../../../../")
	viper.SetConfigType("toml")
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Sprintf("couldn't load config: %s", err))
	}
	GlobalConfig.FeiShuRobotToken = viper.GetString("feishu.accessToken")
}

var GlobalConfig = globalConfig{}

type globalConfig struct {
	FeiShuRobotToken string
}
