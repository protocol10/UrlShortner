package main

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application.
type Config struct {
	Shortener ShortenerConfig `mapstructure:"shortener"`
}

// ShortenerConfig holds configuration specific to URL encoding.
type ShortenerConfig struct {
	HashingAlgorithm string `mapstructure:"hashing_algorithm"`
	MaxCharLimit     int    `mapstructure:"max_char_limit"`
}

// loadConfig reads the config.yaml file and parses it into the Config struct.
func loadConfig() (*Config, error) {
	viper.SetConfigName("config") // name of config file (without extension)
	viper.SetConfigType("yaml")   // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath(".")      // optionally look for config in the working directory

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("unable to decode into struct: %w", err)
	}

	return &config, nil
}

func main() {
	config, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Println("URL Shortener initialized!")
	fmt.Printf("Using Algorithm: %s\n", config.Shortener.HashingAlgorithm)
	fmt.Printf("Max Character Limit: %d\n", config.Shortener.MaxCharLimit)
}
