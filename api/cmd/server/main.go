// api/cmd/server/main.go
package main

import (
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/teclara/gopulse/api/internal/cache"
	"github.com/teclara/gopulse/api/internal/config"
	"github.com/teclara/gopulse/api/internal/handlers"
	"github.com/teclara/gopulse/api/internal/metrolinx"
)

func main() {
	cfg := config.Load()

	client := metrolinx.NewClient(cfg.MetrolinxBaseURL, cfg.MetrolinxAPIKey)
	c := cache.New()
	h := handlers.New(client, c)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", h.Health)
	mux.HandleFunc("GET /api/departures/union", h.UnionDepartures)
	mux.HandleFunc("GET /api/departures/{stopCode}", h.StopDepartures)
	mux.HandleFunc("GET /api/trains", h.Trains)
	mux.HandleFunc("GET /api/trains/positions", h.TrainPositions)
	mux.HandleFunc("GET /api/alerts/service", h.ServiceAlerts)
	mux.HandleFunc("GET /api/alerts/info", h.InfoAlerts)
	mux.HandleFunc("GET /api/exceptions", h.Exceptions)
	mux.HandleFunc("GET /api/schedule/lines/{date}", h.ScheduleLines)
	mux.HandleFunc("GET /api/schedule/journey", h.ScheduleJourney)
	mux.HandleFunc("GET /api/fares/{from}/{to}", h.Fares)
	mux.HandleFunc("GET /api/stops", h.AllStops)
	mux.HandleFunc("GET /api/stops/{code}", h.StopDetails)

	handler := corsMiddleware(cfg.AllowedOrigins, mux)

	slog.Info("starting gopulse-api", "port", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, handler); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}

func corsMiddleware(allowedOrigins string, next http.Handler) http.Handler {
	origins := strings.Split(allowedOrigins, ",")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		for _, o := range origins {
			if strings.TrimSpace(o) == origin {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				break
			}
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
