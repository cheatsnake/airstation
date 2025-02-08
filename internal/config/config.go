package config

import (
	"os"
	"path/filepath"
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
		println(key + " environment variable is not set")
		os.Exit(1)
	}

	if len(secretKey) < minSecretKeyLength {
		println(key + " is too short")
		os.Exit(1)
	}

	return secretKey
}
