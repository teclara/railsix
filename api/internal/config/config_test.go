// api/internal/config/config_test.go
package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/teclara/railsix/api/internal/config"
)

func TestLoad_Defaults(t *testing.T) {
	cfg := config.Load()
	if cfg.Port != "8080" {
		t.Fatalf("expected default port 8080, got %s", cfg.Port)
	}
	if cfg.MetrolinxBaseURL != "https://api.openmetrolinx.com/OpenDataAPI/api/V1" {
		t.Fatalf("unexpected base URL: %s", cfg.MetrolinxBaseURL)
	}
}

func TestLoad_FromEnv(t *testing.T) {
	t.Setenv("PORT", "9090")
	t.Setenv("METROLINX_API_KEY", "test-key")
	t.Setenv("ALLOWED_ORIGINS", "https://example.com")

	cfg := config.Load()
	if cfg.Port != "9090" {
		t.Fatalf("expected port 9090, got %s", cfg.Port)
	}
	if cfg.MetrolinxAPIKey != "test-key" {
		t.Fatalf("expected api key test-key, got %s", cfg.MetrolinxAPIKey)
	}
	if cfg.AllowedOrigins != "https://example.com" {
		t.Fatalf("expected allowed origins, got %s", cfg.AllowedOrigins)
	}
}

func TestLoad_FromDotEnvLocal(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".env.local"), []byte("METROLINX_API_KEY=file-key\nPORT=9191\n"), 0o600); err != nil {
		t.Fatalf("write .env.local: %v", err)
	}

	prev, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir temp dir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(prev)
	})

	cfg := config.Load()
	if cfg.MetrolinxAPIKey != "file-key" {
		t.Fatalf("expected dotenv api key, got %q", cfg.MetrolinxAPIKey)
	}
	if cfg.Port != "9191" {
		t.Fatalf("expected dotenv port 9191, got %s", cfg.Port)
	}
}

func TestLoad_EnvOverridesDotEnvLocal(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".env.local"), []byte("METROLINX_API_KEY=file-key\n"), 0o600); err != nil {
		t.Fatalf("write .env.local: %v", err)
	}

	prev, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir temp dir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(prev)
	})

	t.Setenv("METROLINX_API_KEY", "env-key")

	cfg := config.Load()
	if cfg.MetrolinxAPIKey != "env-key" {
		t.Fatalf("expected env api key to win, got %q", cfg.MetrolinxAPIKey)
	}
}
