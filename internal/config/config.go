package config

import (
	"os"
	"path/filepath"
)

type Config struct {
	TracksDir string
	TmpDir    string
	WebDir    string
	HTTPPort  string
}

func Load() *Config {
	return &Config{
		TracksDir: getEnv("AIRSTATION_TRACKS_DIR", filepath.Join("static", "tracks")),
		TmpDir:    getEnv("AIRSTATION_TMP_DIR", filepath.Join("static", "tmp")),
		WebDir:    getEnv("AIRSTATION_WEB_DIR", filepath.Join("web")),
		HTTPPort:  getEnv("AIRSTATION_HTTP_PORT", "8080"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
