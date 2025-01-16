package main

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

type (
	Config struct {
		Nodes              []string `mapstructure:"nodes"`
		DB                 DB       `mapstructure:"db"`
		Server             Server   `mapstructure:"server"`
		BlockNumber        uint64   `mapstructure:"block_number"`
		ListBillsPageLimit int      `mapstructure:"list_bills_page_limit"`
	}

	DB struct {
		URL string `mapstructure:"url"`
	}

	Server struct {
		Address string `mapstructure:"address"`
	}
)

const envPrefix = "BLOCK_EXPLORER"

func LoadConfig(configFilePath string) (*Config, error) {
	viper.AutomaticEnv()
	viper.SetEnvPrefix(envPrefix)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Attempt to read the config file if provided
	if configFilePath != "" {
		dir, file := filepath.Split(configFilePath)
		ext := filepath.Ext(file)

		viper.AddConfigPath(dir)
		viper.SetConfigName(file[:len(file)-len(ext)])
		viper.SetConfigType(ext[1:])

		if err := viper.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				return nil, fmt.Errorf("failed to read config file: %w", err)
			}
			log.Println("No config file found, using environment variables only.")
		} else {
			log.Printf("Config file %s loaded successfully.", configFilePath)
		}
	} else {
		log.Println("No config file provided, using environment variables only.")
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}
