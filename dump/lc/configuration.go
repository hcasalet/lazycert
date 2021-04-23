package lc

import (
	"log"
	"os"
)

type Config struct {
	configFilePath     string
	privateKeyFileName string
	f                  int32
}

func NewConfig() *Config {
	homePath := os.Getenv("HOME")
	log.Printf("HOME = %v", homePath)
	lcConfigurationPath := homePath + "/.lazycert"
	//var err error
	var _, err = os.Stat(lcConfigurationPath)
	if os.IsNotExist(err) {
		err = os.Mkdir(lcConfigurationPath, 0755)
		if err != nil {
			log.Fatalf("Could not create configuration file directory: %v", err)
		}
	}
	privateKeyFilePath := lcConfigurationPath + "/privateKey_%s.pem"
	c := &Config{configFilePath: lcConfigurationPath, privateKeyFileName: privateKeyFilePath, f: 1}
	log.Printf("Configuration: %v", c)
	return c
}
