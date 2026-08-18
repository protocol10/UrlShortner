package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application.
type Config struct {
	Shortener ShortenerConfig `mapstructure:"shortener"`
	Database  DatabaseConfig  `mapstructure:"database"`
}

// ShortenerConfig holds configuration specific to URL encoding.
type ShortenerConfig struct {
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

// LoadConfig reads the config.yaml file and parses it into the Config struct.
func LoadConfig() (*Config, error) {
	viper.SetConfigName("config") // name of config file (without extension)
	viper.SetConfigType("yaml")   // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath(".")      // optionally look for config in the working directory

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to decode into struct: %w", err)
	}

	return &cfg, nil
}
