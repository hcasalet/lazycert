package lc

import (
	"fmt"
	"log"
	"os"
)

type Config struct {
	ConfigFilePath     string
	PrivateKeyFileName string
	F                  int
}

func NewConfig(prefix string) *Config {
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
	privateKeyFilePath = fmt.Sprintf(privateKeyFilePath, "TE")
	c := &Config{ConfigFilePath: lcConfigurationPath, PrivateKeyFileName: privateKeyFilePath, F: 1}
	log.Printf("Configuration: %v", c)
	return c
}
