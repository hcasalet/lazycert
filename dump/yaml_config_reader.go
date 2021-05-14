package main

import (
	"fmt"
	"github.com/spf13/viper"
)

type LCConfig struct {
	viper *viper.Viper
}

func NewLCConfig() *LCConfig {
	viper.SetConfigName("config.yml")
	viper.AddConfigPath("./dump/")
	viper.SetConfigType("yaml")
	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("Error reading configuration file: %v", err)
		panic("Cannot read configuration file.")
	} else {
		fmt.Println("Config file read.")
		myviper := &LCConfig{
			viper: viper.GetViper(),
		}
		return myviper
	}
}
