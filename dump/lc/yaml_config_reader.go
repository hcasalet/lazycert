package lc

import (
	"fmt"
	"github.com/spf13/viper"
)

type YamlConfig struct {
	Viper *viper.Viper
}

func NewYamlConfig() *YamlConfig {
	viper.SetConfigName("config.yml")
	viper.AddConfigPath("./")
	viper.SetConfigType("yaml")
	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("Error reading configuration file: %v", err)
		panic("Cannot read configuration file.")
	} else {
		fmt.Println("Config file read.")
		myviper := &YamlConfig{
			Viper: viper.GetViper(),
		}
		return myviper
	}
}
