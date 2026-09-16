package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application.
type Config struct {
	Shortener ShortenerConfig `mapstructure:"shortener"`
	Database  DatabaseConfig  `mapstructure:"database"`
	Redis     RedisConfig     `mapstructure:"redis"`
	Server    ServerConfig    `mapstructure:"server"`
}

// ShortenerConfig holds configuration specific to URL encoding.
type ShortenerConfig struct {
	Approach              string `mapstructure:"approach"`
	HashingAlgorithm      string `mapstructure:"hashing_algorithm"`
	MaxCharLimit          int    `mapstructure:"max_char_limit"`
	ObfuscationMultiplier uint64 `mapstructure:"obfuscation_multiplier"`
	ObfuscationXOR        uint64 `mapstructure:"obfuscation_xor"`
}

// DatabaseConfig holds PostgreSQL configuration.
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
}

// RedisConfig holds Redis configuration.
type RedisConfig struct {
	Host       string `mapstructure:"host"`
	Port       int    `mapstructure:"port"`
	Password   string `mapstructure:"password"`
	CounterKey string `mapstructure:"counter_key"`
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
	if cfg.Redis.Host == "" {
		cfg.Redis.Host = "127.0.0.1"
	}
	if cfg.Redis.Port == 0 {
		cfg.Redis.Port = 6379
	}
	if cfg.Redis.CounterKey == "" {
		cfg.Redis.CounterKey = "url_shortener:counter"
	}
	if cfg.Shortener.ObfuscationMultiplier == 0 {
		cfg.Shortener.ObfuscationMultiplier = 11400714819323198485
	}
	if cfg.Shortener.ObfuscationXOR == 0 {
		cfg.Shortener.ObfuscationXOR = 0x5BF03635467C061D
	}

	return &cfg, nil
}
