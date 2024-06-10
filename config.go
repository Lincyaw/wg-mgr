package main

import (
	"gopkg.in/yaml.v3"
	"os"
)

func LoadServerConfig(filePath string) (*ServerConfig, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var config ServerConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
