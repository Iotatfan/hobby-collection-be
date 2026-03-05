package config

import (
	"os"
	"strings"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

type ServerConfig struct {
	Port int `mapstructure:"port" yaml:"port"`
}

type PostgresConfig struct {
	Host     string `mapstructure:"host" yaml:"host"`
	Port     int    `mapstructure:"port" yaml:"port"`
	User     string `mapstructure:"user" yaml:"user"`
	Password string `mapstructure:"password" yaml:"password"`
	DBName   string `mapstructure:"dbname" yaml:"dbname"`
}

type JWTConfig struct {
	Secret  string `mapstructure:"secret" yaml:"secret"`
	Access  int    `mapstructure:"access" yaml:"access"`
	Refresh int    `mapstructure:"refresh" yaml:"refresh"`
}

type CloudinaryConfig struct {
	Name   string `mapstructure:"name" yaml:"name"`
	Key    string `mapstructure:"key" yaml:"key"`
	Secret string `mapstructure:"secret" yaml:"secret"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host" yaml:"host"`
	Password string `mapstructure:"password" yaml:"password"`
	Database int    `mapstructure:"database" yaml:"database"`
	Port     int    `mapstructure:"port" yaml:"port"`
}

type Config struct {
	Server     ServerConfig     `mapstructure:"server" yaml:"server"`
	Postgres   PostgresConfig   `mapstructure:"postgres" yaml:"postgres"`
	JWT        JWTConfig        `mapstructure:"jwt" yaml:"jwt"`
	Cloudinary CloudinaryConfig `mapstructure:"cloudinary" yaml:"cloudinary"`
	Redis      RedisConfig      `mapstructure:"redis" yaml:"redis"`
}

var AppConfig *Config

func checkConfig(cfg *Config) {
	if cfg.Postgres.Host == "" {
		panic("host name not configured")
	}
}

func checkConfigFile() error {
	if _, err := os.Stat("config.yml"); err == nil {
		return nil // file already exists
	}

	var cfg Config

	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return err
	}

	return os.WriteFile("config.yml", data, 0644)
}

func InitConfig() error {
	checkConfigFile()

	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath(".")

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		return err
	}

	var cfg Config
	err = viper.Unmarshal(&cfg)
	if err != nil {
		return err
	}

	AppConfig = &cfg

	return nil
}

func GetConfig() *Config {
	return AppConfig
}
