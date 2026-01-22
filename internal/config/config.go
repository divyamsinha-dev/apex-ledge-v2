package config

import (
	"os"
	"strconv"
)

type Config struct {
	DBURL       string
	GRPCPort    string
	JWTSecret   string
	WorkerCount int
}

func Load() *Config {
	return &Config{
		DBURL:       getEnv("DB_URL", "postgres://user:pass@localhost:5432/ledger?sslmode=disable"),
		GRPCPort:    getEnv("GRPC_PORT", "50051"),
		JWTSecret:   getEnv("JWT_SECRET", "production-secret-key"),
		WorkerCount: getEnvInt("WORKER_COUNT", 5),
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	v := getEnv(key, "")
	if i, err := strconv.Atoi(v); err == nil {
		return i
	}
	return fallback
}
