package config

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

const minSecretLength = 10

type Config struct {
	DBDir        string
	DBFile       string
	TracksDir    string
	TmpDir       string
	PlayerDir    string
	StudioDir    string
	HTTPPort     string
	JWTSign      string
	SecretKey    string
	SecureCookie bool
}

func Load() *Config {
	_ = godotenv.Load() // For development

	return &Config{
		DBDir:        getEnv("AIRSTATION_DB_DIR", filepath.Join("storage")),
		DBFile:       getEnv("AIRSTATION_DB_FILE", "storage.db"),
		TracksDir:    getEnv("AIRSTATION_TRACKS_DIR", filepath.Join("static", "tracks")),
		TmpDir:       getEnv("AIRSTATION_TMP_DIR", filepath.Join("static", "tmp")),
		PlayerDir:    getEnv("AIRSTATION_PLAYER_DIR", filepath.Join("player")),
		StudioDir:    getEnv("AIRSTATION_STUDIO_DIR", filepath.Join("studio")),
		HTTPPort:     getEnv("AIRSTATION_HTTP_PORT", "7331"),
		JWTSign:      getSecret("AIRSTATION_JWT_SIGN"),
		SecretKey:    getSecret("AIRSTATION_SECRET_KEY"),
		SecureCookie: getEnvBool("AIRSTATION_SECURE_COOKIE", false),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}

	val = strings.ToLower(val)
	return val == "1" || val == "true" || val == "yes" || val == "on"
}

func getSecret(key string) string {
	secretKey := os.Getenv(key)

	if secretKey == "" {
		log.Fatal(key + " environment variable is not set")
	}

	if len(secretKey) < minSecretLength {
		log.Fatal(key + " is too short")
	}

	return secretKey
}
