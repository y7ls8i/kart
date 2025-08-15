package config_test

import (
	"testing"

	"github.com/y7ls8i/kart/config"
)

func TestConfig(t *testing.T) {
	conf := config.ReadConfig("../config.toml")
	if conf == nil {
		t.Fatalf("config is nil")
	}
	if conf.Server.Listen == "" {
		t.Fatalf("config.Server.Listen is empty")
	}
	if conf.MongoDB.URI == "" {
		t.Fatalf("config.MongoDB.URI is empty")
	}
}
