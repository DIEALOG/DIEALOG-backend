package config

import (
	"log"
	"os"

	"github.com/spf13/viper"
)

var E Env

type Env struct {
	AppEnv                 string `mapstructure:"APP_ENV"`
	ServerAddress          string `mapstructure:"SERVER_ADDRESS"`
	ContextTimeout         int    `mapstructure:"CONTEXT_TIMEOUT"`
	DBHost                 string `mapstructure:"DB_HOST"`
	DBPort                 string `mapstructure:"DB_PORT"`
	DBUser                 string `mapstructure:"DB_USER"`
	DBPass                 string `mapstructure:"DB_PASS"`
	DBName                 string `mapstructure:"DB_NAME"`
	AccessTokenExpiryHour  int    `mapstructure:"ACCESS_TOKEN_EXPIRY_HOUR"`
	RefreshTokenExpiryHour int    `mapstructure:"REFRESH_TOKEN_EXPIRY_HOUR"`
	AccessTokenSecret      string `mapstructure:"ACCESS_TOKEN_SECRET"`
	RefreshTokenSecret     string `mapstructure:"REFRESH_TOKEN_SECRET"`
	InternalSecretKey      string `mapstructure:"INTERNAL_SECRET_KEY"`
	InternalCryptoUrl      string `mapstructure:"INTERNAL_CRYPTO_URL"`
}

var envKeys = []string{
	"APP_ENV",
	"SERVER_ADDRESS",
	"CONTEXT_TIMEOUT",
	"DB_HOST",
	"DB_PORT",
	"DB_USER",
	"DB_PASS",
	"DB_NAME",
	"ACCESS_TOKEN_EXPIRY_HOUR",
	"REFRESH_TOKEN_EXPIRY_HOUR",
	"ACCESS_TOKEN_SECRET",
	"REFRESH_TOKEN_SECRET",
	"INTERNAL_SECRET_KEY",
	"INTERNAL_CRYPTO_URL",
}

func NewEnv() {
	viper.Reset()
	viper.SetConfigFile(".env")

	if _, err := os.Stat(".env"); err == nil {
		if err := viper.ReadInConfig(); err != nil {
			log.Fatal("Environment can't be loaded from .env: ", err)
		}
	} else if !os.IsNotExist(err) {
		log.Fatal("Can't access .env: ", err)
	}

	// Environment variables take precedence over values from .env, so the same
	// binary works locally (.env) and in containers (environment variables).
	for _, key := range envKeys {
		if err := viper.BindEnv(key); err != nil {
			log.Fatal("Environment can't bind ", key, ": ", err)
		}
	}

	if err := viper.Unmarshal(&E); err != nil {
		log.Fatal("Environment can't be loaded: ", err)
	}

	if E.AppEnv == "development" {
		log.Println("The App is running in development env")
	}
}
