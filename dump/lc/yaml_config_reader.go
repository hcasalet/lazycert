package lc

import (
	"fmt"
	"github.com/spf13/viper"
)

type LCConfig struct {
	Viper *viper.Viper
}

func NewLCConfig() *LCConfig {
	viper.SetConfigName("config.yml")
	viper.AddConfigPath("./")
	viper.SetConfigType("yaml")
	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("Error reading configuration file: %v", err)
		panic("Cannot read configuration file.")
	} else {
		fmt.Println("Config file read.")
		myviper := &LCConfig{
			Viper: viper.GetViper(),
		}
		return myviper
	}
}
