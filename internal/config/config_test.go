package config

import "testing"

func TestConfigValidateRejectsInvalidPort(t *testing.T) {
	cfg := Config{
		Env:          "prod",
		Port:         "not-a-port",
		AssetsPath:   "web/assets",
		DatabasePath: "data/processed/gouv-viz.sqlite",
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want invalid port error")
	}
}

func TestConfigAddr(t *testing.T) {
	cfg := Config{Port: "9456"}

	if got := cfg.Addr(); got != ":9456" {
		t.Fatalf("Addr() = %q, want :9456", got)
	}
}
