package config

import (
	"strconv"

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

func InitConfig() *Config {
	// viper.SetConfigName("config")
	// viper.SetConfigType("yml")
	// viper.AddConfigPath(".")

	viper.SetConfigFile("ENV")
	viper.ReadInConfig()
	viper.AutomaticEnv()

	// viper.SetConfigFile(".env")
	// viper.AutomaticEnv()

	// err := viper.ReadInConfig()
	// if err != nil {
	// 	panic(fmt.Errorf("fatal error config file: %w", err))
	// }

	var cfg Config

	// err = viper.Unmarshal(&cfg)
	// if err != nil {
	// 	panic(fmt.Errorf("fatal error config file: %w", err))
	// }

	cfg.Postgres.Host = viper.GetString("DB_HOST")
	cfg.Postgres.Port, _ = strconv.Atoi(viper.GetString("DB_PORT"))
	cfg.Postgres.User = viper.GetString("DB_USER")
	cfg.Postgres.Password = viper.GetString("DB_PASSWORD")
	cfg.Postgres.DBName = viper.GetString("DB_NAME")

	checkConfig(&cfg)
	return &cfg
}
