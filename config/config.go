package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application.
type Config struct {
	Shortener ShortenerConfig `mapstructure:"shortener"`
	Database  DatabaseConfig  `mapstructure:"database"`
	Server    ServerConfig    `mapstructure:"server"`
}

// ShortenerConfig holds configuration specific to URL encoding.
type ShortenerConfig struct {
	Approach         string `mapstructure:"approach"`
	HashingAlgorithm string `mapstructure:"hashing_algorithm"`
	MaxCharLimit     int    `mapstructure:"max_char_limit"`
}

// DatabaseConfig holds PostgreSQL configuration.
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
}

// ServerConfig holds HTTP server configuration.
type ServerConfig struct {
	Port int `mapstructure:"port"`
}

// LoadConfig reads the config.yaml file and parses it into the Config struct.
func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to decode into struct: %w", err)
	}

	// Set defaults
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}

	return &cfg, nil
}
