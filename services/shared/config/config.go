package config

import (
	"fmt"
	"os"
	"strconv"
)

// Environment variable name constants.
const (
	EnvNATSURL         = "NATS_URL"
	EnvRedisAddr       = "REDIS_ADDR"
	EnvRedisPassword   = "REDIS_PASSWORD"
	EnvMetrolinxAPIKey = "METROLINX_API_KEY"
	EnvMetrolinxBase   = "METROLINX_BASE_URL"
	EnvPort            = "PORT"
	EnvAllowedOrigins  = "ALLOWED_ORIGINS"
	EnvGTFSStaticURL   = "GTFS_STATIC_URL"
	EnvGTFSStaticAddr  = "GTFS_STATIC_ADDR"
)

// Default values for environment variables.
const (
	DefaultNATSURL        = "nats://localhost:4222"
	DefaultRedisAddr      = "localhost:6379"
	DefaultMetrolinxBase  = "https://api.openmetrolinx.com/OpenDataAPI/api/V1"
	DefaultGTFSStaticURL  = "https://assets.metrolinx.com/raw/upload/Documents/Metrolinx/Open%20Data/GO-GTFS.zip"
	DefaultGTFSStaticAddr = "http://localhost:8081"
)

// EnvOr returns the value of the environment variable named by key,
// or fallback if the variable is not set or empty.
func EnvOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// EnvOrInt returns the parsed integer value of the environment variable named
// by key, or fallback if the variable is not set, invalid, or non-positive.
func EnvOrInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(v)
	if err != nil || parsed <= 0 {
		return fallback
	}

	return parsed
}

// Require returns the value of the environment variable named by key,
// or an error if the variable is not set or empty.
func Require(key string) (string, error) {
	v := os.Getenv(key)
	if v == "" {
		return "", fmt.Errorf("required environment variable %s is not set", key)
	}
	return v, nil
}
