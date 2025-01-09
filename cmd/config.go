package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type (
	Config struct {
		Nodes []struct {
			URL string `yaml:"url"`
		} `yaml:"nodes"`

		DB struct {
			URL string `yaml:"url"` // MongoDB connection string
		} `yaml:"db"`

		Server struct {
			Address string `yaml:"address"`
		} `yaml:"server"`

		BlockNumber        uint64 `yaml:"blockNumber"`
		ListBillsPageLimit int
	}
)

func ReadConfig(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}

	var config Config
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode config file: %w", err)
	}

	return &config, nil
}
