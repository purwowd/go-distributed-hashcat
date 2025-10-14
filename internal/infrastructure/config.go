package infrastructure

import (
	"strings"

	"github.com/spf13/viper"
)

func init() {
	viper.SetConfigFile(".env")
	_ = viper.ReadInConfig()

	viper.AutomaticEnv()
}

func GlobalXToken() string {
	return strings.TrimSpace(viper.GetString("GLOBAL_X_TOKEN"))
}
