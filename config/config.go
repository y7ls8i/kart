// Package config is http server configuration.
package config

import (
	"log/slog"
	"os"

	"github.com/BurntSushi/toml"
)

// Server structure
type Server struct {
	Mode     string
	Listen   string
	Certfile string
	Keyfile  string
}

// MongoDB structure.
type MongoDB struct {
	URI string
	DB  string
}

// Config structure
type Config struct {
	Server  Server
	MongoDB MongoDB
}

// ReadConfig from a config file
func ReadConfig(configFile string) (config *Config) {
	config = &Config{}
	if _, err := toml.DecodeFile(configFile, &config); err != nil {
		slog.Error("Error parsing config", "error", err)
		os.Exit(1)
	}
	return config
}
