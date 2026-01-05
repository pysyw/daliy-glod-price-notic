package cfg

import (
	"os"

	"github.com/joho/godotenv"
)

func init() {
	godotenv.Load()
	// viper.SetConfigName("config")
	// viper.AddConfigPath(".")
	// viper.AddConfigPath(".")
	// viper.AddConfigPath("../.")
	// viper.AddConfigPath("../../.")
	// viper.AddConfigPath("../../../.")
	// viper.AddConfigPath("../../../../../.")
	// viper.AddConfigPath("../../../../")
	// viper.SetConfigType("toml")
	// if err := viper.ReadInConfig(); err != nil {
	// 	panic(fmt.Sprintf("couldn't load config: %s", err))
	// }
	// GlobalConfig.FeiShuRobotToken = viper.GetString("feishu.accessToken")
	GlobalConfig.FeiShuRobotToken = os.Getenv("FEI_SHU_ACCESS_TOKEN")
}

var GlobalConfig = globalConfig{}

type globalConfig struct {
	FeiShuRobotToken string
}
