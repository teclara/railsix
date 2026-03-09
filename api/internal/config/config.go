// api/internal/config/config.go
package config

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Port             string
	MetrolinxAPIKey  string
	MetrolinxBaseURL string
	AllowedOrigins   string
	GTFSStaticURL    string
}

func Load() Config {
	fileEnv := loadDotEnv()

	return Config{
		Port:             envOrFile("PORT", "8080", fileEnv),
		MetrolinxAPIKey:  envOrFile("METROLINX_API_KEY", "", fileEnv),
		MetrolinxBaseURL: envOrFile("METROLINX_BASE_URL", "https://api.openmetrolinx.com/OpenDataAPI/api/V1", fileEnv),
		AllowedOrigins:   envOrFile("ALLOWED_ORIGINS", "http://localhost:5173", fileEnv),
		GTFSStaticURL:    envOrFile("GTFS_STATIC_URL", "https://assets.metrolinx.com/raw/upload/Documents/Metrolinx/Open%20Data/GO-GTFS.zip", fileEnv),
	}
}

func envOrFile(key, fallback string, fileEnv map[string]string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	if v := fileEnv[key]; v != "" {
		return v
	}
	return fallback
}

func loadDotEnv() map[string]string {
	cwd, err := os.Getwd()
	if err != nil {
		return map[string]string{}
	}

	paths := []string{
		filepath.Join(cwd, "..", ".env"),
		filepath.Join(cwd, ".env"),
		filepath.Join(cwd, "..", ".env.local"),
		filepath.Join(cwd, ".env.local"),
	}

	env := make(map[string]string)
	for _, path := range paths {
		loadDotEnvFile(path, env)
	}
	return env
}

func loadDotEnvFile(path string, env map[string]string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		line = strings.TrimPrefix(line, "export ")
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if len(value) >= 2 {
			if (value[0] == '"' && value[len(value)-1] == '"') ||
				(value[0] == '\'' && value[len(value)-1] == '\'') {
				value = value[1 : len(value)-1]
			}
		}
		if key != "" {
			env[key] = value
		}
	}
}
