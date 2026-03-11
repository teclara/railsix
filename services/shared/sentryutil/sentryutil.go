package sentryutil

import (
	"log/slog"
	"os"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/teclara/railsix/shared/config"
)

// Init initializes the Sentry SDK if SENTRY_DSN is set.
// Returns true if Sentry was initialized, false if skipped (no DSN).
// The caller should defer Flush() if Init returns true.
func Init(serviceName string) bool {
	dsn := os.Getenv(config.EnvSentryDSN)
	if dsn == "" {
		slog.Info("SENTRY_DSN not set, Sentry disabled")
		return false
	}

	err := sentry.Init(sentry.ClientOptions{
		Dsn:              dsn,
		Environment:      config.EnvOr("ENVIRONMENT", "production"),
		Release:          serviceName,
		TracesSampleRate: 0.2,
		EnableTracing:    true,
	})
	if err != nil {
		slog.Error("sentry.Init failed", "error", err)
		return false
	}

	slog.Info("Sentry initialized", "service", serviceName)
	return true
}

// Flush drains buffered Sentry events. Call with defer after Init.
func Flush() {
	sentry.Flush(2 * time.Second)
}
