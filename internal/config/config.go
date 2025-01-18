package config

import (
	"path/filepath"
	"sync"
)

type Config struct {
	TracksDir string
	TmpDir    string
	WebDir    string
	HTTPPort  string
}

var (
	configInstance *Config
	once           sync.Once
)

func GetConfig() *Config {
	once.Do(func() {
		configInstance = &Config{
			TracksDir: filepath.Join("static", "tracks"),
			TmpDir:    filepath.Join("static", "tmp"),
			WebDir:    filepath.Join("web"),
			HTTPPort:  "8080",
		}
	})
	return configInstance
}
