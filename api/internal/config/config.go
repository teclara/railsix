// api/internal/config/config.go
package config

import "os"

type Config struct {
	Port             string
	MetrolinxAPIKey  string
	MetrolinxBaseURL string
	AllowedOrigins   string
}

func Load() Config {
	return Config{
		Port:             envOr("PORT", "8080"),
		MetrolinxAPIKey:  os.Getenv("METROLINX_API_KEY"),
		MetrolinxBaseURL: envOr("METROLINX_BASE_URL", "https://api.openmetrolinx.com/OpenDataAPI/api/V1"),
		AllowedOrigins:   envOr("ALLOWED_ORIGINS", "http://localhost:5173"),
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
