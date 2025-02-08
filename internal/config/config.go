package config

import (
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

const minSecretKeyLength = 10

type Config struct {
	TracksDir string
	TmpDir    string
	WebDir    string
	HTTPPort  string
	SecretKey string
}

func Load() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Parsing .env file failed: " + err.Error())
	}

	return &Config{
		TracksDir: getEnv("AIRSTATION_TRACKS_DIR", filepath.Join("static", "tracks")),
		TmpDir:    getEnv("AIRSTATION_TMP_DIR", filepath.Join("static", "tmp")),
		WebDir:    getEnv("AIRSTATION_WEB_DIR", filepath.Join("web")),
		HTTPPort:  getEnv("AIRSTATION_HTTP_PORT", "8080"),
		SecretKey: getSecretKey("AIRSTATION_SECRET_KEY"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getSecretKey(key string) string {
	secretKey := os.Getenv(key)

	if secretKey == "" {
		log.Fatal(key + " environment variable is not set")
	}

	if len(secretKey) < minSecretKeyLength {
		log.Fatal(key + " is too short")
	}

	return secretKey
}
