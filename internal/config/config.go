package config

import (
	"strings"

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
	Redis      RedisConfig
}

func checkConfig(cfg *Config) {
	if cfg.Postgres.Host == "" {
		panic("host name not configured")
	}
}

func InitConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath(".")

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = viper.Unmarshal(&cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
