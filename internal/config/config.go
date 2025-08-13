package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type ServerConfig struct {
	Port int
}

type PostgresConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

type JWTConfig struct {
	Secret  string
	Access  int
	Refresh int
}

type CloudinaryConfig struct {
	Name   string
	Key    string
	Secret string
}

type GomailConfig struct {
	Name     string
	Host     string
	Port     int
	Email    string
	Password string
}

type RedisConfig struct {
	Host     string
	Password string
	Database int
	Port     int
}
type Config struct {
	Server     ServerConfig
	Postgres   PostgresConfig
	JWT        JWTConfig
	Cloudinary CloudinaryConfig
	Gomail     GomailConfig
	Redis      RedisConfig
}

func checkConfig(cfg *Config) {
	if cfg.Postgres.Host == "" {
		panic("host name not configured")
	}
}

func InitConfig() *Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	var cfg Config

	err = viper.Unmarshal(&cfg)
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	checkConfig(&cfg)
	return &cfg
}
