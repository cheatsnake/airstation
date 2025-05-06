package config

import (
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

const minSecretLength = 10

type Config struct {
	DBFile    string
	TracksDir string
	TmpDir    string
	PlayerDir string
	StudioDir string
	HTTPPort  string
	JWTSign   string
	SecretKey string
}

func Load() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Parsing .env file failed: " + err.Error())
	}

	return &Config{
		DBFile:    getEnv("AIRSTATION_DB_FILE", filepath.Join("storage", "storage.db")),
		TracksDir: getEnv("AIRSTATION_TRACKS_DIR", filepath.Join("static", "tracks")),
		TmpDir:    getEnv("AIRSTATION_TMP_DIR", filepath.Join("static", "tmp")),
		PlayerDir: getEnv("AIRSTATION_PLAYER_DIR", filepath.Join("player")),
		StudioDir: getEnv("AIRSTATION_STUDIO_DIR", filepath.Join("studio")),
		HTTPPort:  getEnv("AIRSTATION_HTTP_PORT", "7331"),
		JWTSign:   getSecret("AIRSTATION_JWT_SIGN"),
		SecretKey: getSecret("AIRSTATION_SECRET_KEY"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
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
