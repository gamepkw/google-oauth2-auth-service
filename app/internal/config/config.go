package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

func InitializeViper() {

	viper.SetConfigName("config")
	env := os.Getenv("DOCKER_ENV")
	if env == "docker" {
		viper.AddConfigPath("../config")
	} else {
		viper.AddConfigPath("../../config")
	}
	viper.AutomaticEnv()
	viper.SetConfigType("yml")
	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("Error reading config file, %s", err)
	}
}
