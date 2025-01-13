package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type (
	Config struct {
		Nodes []Node `yaml:"nodes"`

		DB DB `yaml:"db"`

		Server Server `yaml:"server"`

		BlockNumber        uint64 `yaml:"blockNumber"`
		ListBillsPageLimit int
	}

	Node struct {
		URL string `yaml:"url"`
	}

	DB struct {
		URL string `yaml:"url"` // MongoDB connection string
	}

	Server struct {
		Address string `yaml:"address"`
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
