package handlers

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func Init() {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("[init] .env not loaded; using environment/default values")
	}

	setDefaultEnv("ENV", "dev")
	setDefaultEnv("PORT", "9456")
	setDefaultEnv("ASSETS_PATH", "web/assets")
}

func setDefaultEnv(key string, value string) {
	if os.Getenv(key) == "" {
		_ = os.Setenv(key, value)
	}
}
