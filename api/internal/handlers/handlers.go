package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"regexp"
	"time"

	"github.com/teclara/gopulse/api/internal/cache"
)

var (
	stopCodeRe = regexp.MustCompile(`^[A-Za-z0-9]{2,10}$`)
	dateRe     = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	timeRe     = regexp.MustCompile(`^\d{2}:\d{2}$`)
	maxJournRe = regexp.MustCompile(`^\d{1,2}$`)
)

type Fetcher interface {
	Fetch(ctx context.Context, path string) ([]byte, error)
}

type Handlers struct {
	fetcher Fetcher
	cache   *cache.Cache
}

func New(fetcher Fetcher, cache *cache.Cache) *Handlers {
	return &Handlers{fetcher: fetcher, cache: cache}
}

func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

func writeJSON(w http.ResponseWriter, status int, data []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := w.Write(data); err != nil {
		slog.Warn("write response failed", "error", err)
	}
}

func jsonError(w http.ResponseWriter, msg string, status int) {
	writeJSON(w, status, []byte(`{"error":"`+msg+`"}`))
}

func (h *Handlers) cachedProxy(w http.ResponseWriter, r *http.Request, metrolinxPath string, ttl time.Duration) {
	if data, ok := h.cache.Get(metrolinxPath); ok {
		w.Header().Set("X-Cache", "HIT")
		writeJSON(w, http.StatusOK, data)
		return
	}

	data, err := h.fetcher.Fetch(r.Context(), metrolinxPath)
	if err != nil {
		slog.Error("metrolinx fetch failed", "path", metrolinxPath, "error", err)
		if stale, ok := h.cache.GetStale(metrolinxPath); ok {
			w.Header().Set("X-Cache", "STALE")
			w.Header().Set("X-Cache-Stale", "true")
			writeJSON(w, http.StatusOK, stale)
			return
		}
		jsonError(w, "upstream unavailable", http.StatusBadGateway)
		return
	}

	h.cache.Set(metrolinxPath, data, ttl)
	w.Header().Set("X-Cache", "MISS")
	writeJSON(w, http.StatusOK, data)
}

func (h *Handlers) UnionDepartures(w http.ResponseWriter, r *http.Request) {
	h.cachedProxy(w, r, "/ServiceUpdate/UnionDepartures/All", 30*time.Second)
}

func (h *Handlers) StopDepartures(w http.ResponseWriter, r *http.Request) {
	stopCode := r.PathValue("stopCode")
	if !stopCodeRe.MatchString(stopCode) {
		jsonError(w, "invalid stop code", http.StatusBadRequest)
		return
	}
	h.cachedProxy(w, r, "/Stop/NextService/"+stopCode, 30*time.Second)
}

func (h *Handlers) Trains(w http.ResponseWriter, r *http.Request) {
	h.cachedProxy(w, r, "/ServiceataGlance/Trains/All", 30*time.Second)
}

func (h *Handlers) TrainPositions(w http.ResponseWriter, r *http.Request) {
	h.cachedProxy(w, r, "/Gtfs/Feed/VehiclePosition", 15*time.Second)
}

func (h *Handlers) ServiceAlerts(w http.ResponseWriter, r *http.Request) {
	h.cachedProxy(w, r, "/ServiceUpdate/ServiceAlert/All", 60*time.Second)
}

func (h *Handlers) InfoAlerts(w http.ResponseWriter, r *http.Request) {
	h.cachedProxy(w, r, "/ServiceUpdate/InformationAlert/All", 60*time.Second)
}

func (h *Handlers) Exceptions(w http.ResponseWriter, r *http.Request) {
	h.cachedProxy(w, r, "/ServiceUpdate/Exceptions/All", 60*time.Second)
}

func (h *Handlers) ScheduleLines(w http.ResponseWriter, r *http.Request) {
	date := r.PathValue("date")
	if !dateRe.MatchString(date) {
		jsonError(w, "invalid date format, use YYYY-MM-DD", http.StatusBadRequest)
		return
	}
	h.cachedProxy(w, r, "/Schedule/Line/All/"+date, time.Hour)
}

func (h *Handlers) ScheduleJourney(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	date := q.Get("date")
	from := q.Get("from")
	to := q.Get("to")
	startTime := q.Get("startTime")
	maxJourney := q.Get("maxJourney")
	if maxJourney == "" {
		maxJourney = "3"
	}

	if !dateRe.MatchString(date) || !stopCodeRe.MatchString(from) || !stopCodeRe.MatchString(to) || !timeRe.MatchString(startTime) || !maxJournRe.MatchString(maxJourney) {
		jsonError(w, "invalid query parameters", http.StatusBadRequest)
		return
	}

	path := "/Schedule/Journey/" + date + "/" + from + "/" + to + "/" + startTime + "/" + maxJourney
	h.cachedProxy(w, r, path, 5*time.Minute)
}

func (h *Handlers) Fares(w http.ResponseWriter, r *http.Request) {
	from := r.PathValue("from")
	to := r.PathValue("to")
	if !stopCodeRe.MatchString(from) || !stopCodeRe.MatchString(to) {
		jsonError(w, "invalid station code", http.StatusBadRequest)
		return
	}
	h.cachedProxy(w, r, "/Fares/"+from+"/"+to, 24*time.Hour)
}

func (h *Handlers) AllStops(w http.ResponseWriter, r *http.Request) {
	h.cachedProxy(w, r, "/Stop/All", 24*time.Hour)
}

func (h *Handlers) StopDetails(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	if !stopCodeRe.MatchString(code) {
		jsonError(w, "invalid stop code", http.StatusBadRequest)
		return
	}
	h.cachedProxy(w, r, "/Stop/Details/"+code, 24*time.Hour)
}
