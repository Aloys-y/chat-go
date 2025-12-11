package config

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Auth     AuthConfig
	WebRTC   WebRTCConfig
}

type ServerConfig struct {
	HTTPPort  int `mapstructure:"http_port"`
	GRPCPort  int `mapstructure:"grpc_port"`
	WSPort    int `mapstructure:"ws_port"`
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string `mapstructure:"dbname"`
	Charset  string
	ParseTime bool `mapstructure:"parseTime"`
}

type AuthConfig struct {
	SecretKey  string `mapstructure:"secret_key"`
	TokenExpiry string `mapstructure:"token_expiry"`
}

type WebRTCConfig struct {
	ICEServers []ICEServer `mapstructure:"ice_servers"`
}

type ICEServer struct {
	URLs string `mapstructure:"urls"`
}

var AppConfig Config

// LoadConfig loads configuration from config.yaml file
func LoadConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	if err := viper.Unmarshal(&AppConfig); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	log.Println("Config loaded successfully")
	return nil
}