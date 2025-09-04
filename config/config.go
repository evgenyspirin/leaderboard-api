package config

import (
	"os"
)

type APP struct {
	Name string
	Host string
	Port string
}

type Config struct {
	App APP
}

func getEnv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

func Load() Config {
	app := APP{
		Name: getEnv("SERVICE_NAME", ""),
		Host: getEnv("SERVICE_HOST", ""),
		Port: getEnv("SERVICE_PORT", ""),
	}

	return Config{
		App: app,
	}
}
