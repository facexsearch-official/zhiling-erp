package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig `mapstructure:"server"`
	JWT      JWTConfig    `mapstructure:"jwt"`
	Database DBConfig     `mapstructure:"database"`
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type JWTConfig struct {
	Secret      string `mapstructure:"secret"`
	ExpireHours int    `mapstructure:"expire_hours"`
}

type DBConfig struct {
	DSN          string `mapstructure:"dsn"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
}

var GlobalConfig Config

func Load(path string) error {
	viper.SetConfigFile(path)
	if err := viper.ReadInConfig(); err != nil {
		return err
	}
	return viper.Unmarshal(&GlobalConfig)
}
