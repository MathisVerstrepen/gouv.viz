package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Env          string
	Port         string
	BaseURL      string
	AssetsPath   string
	DatabasePath string
}

func Load() (Config, error) {
	loadEnvFile()

	cfg := Config{
		Env:          envOrDefault("ENV", "dev"),
		Port:         envOrDefault("PORT", "9456"),
		BaseURL:      envOrDefault("BASE_URL", ""),
		AssetsPath:   envOrDefault("ASSETS_PATH", "web/assets"),
		DatabasePath: envOrDefault("DATABASE_PATH", "data/processed/gouv-viz.sqlite"),
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func loadEnvFile() {
	if envFile := os.Getenv("ENV_FILE"); envFile != "" {
		loadEnvFilePath(envFile)
		return
	}

	if _, err := os.Stat(".env"); err == nil {
		loadEnvFilePath(".env")
		return
	}

	loadEnvFilePath("dev.env")
}

func loadEnvFilePath(path string) {
	if err := godotenv.Load(path); err != nil {
		log.Printf("[init] %s not loaded; using environment/default values", path)
	}
}

func (c Config) Validate() error {
	if c.Env == "" {
		return fmt.Errorf("ENV is required")
	}
	if c.AssetsPath == "" {
		return fmt.Errorf("ASSETS_PATH is required")
	}
	if c.DatabasePath == "" {
		return fmt.Errorf("DATABASE_PATH is required")
	}
	port, err := strconv.Atoi(c.Port)
	if err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("PORT must be a valid TCP port")
	}
	return nil
}

func (c Config) Addr() string {
	return ":" + c.Port
}

func (c Config) IsProd() bool {
	return c.Env == "prod"
}

func envOrDefault(key string, value string) string {
	if current := os.Getenv(key); current != "" {
		return current
	}
	return value
}
