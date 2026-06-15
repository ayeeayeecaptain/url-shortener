package config

import "os"

type Config struct {
	ServerPort  string
	DatabaseURL string
	RedisURL    string
}

func Load() *Config {
	return &Config{
		ServerPort:  getEnv("SERVER_PORT", ":8080"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5433/shortener?sslmode=disable"),
		RedisURL:    getEnv("REDIS_URL", "localhost:6379"),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
