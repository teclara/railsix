package gtfs

import (
	"context"
	"log/slog"
	"math"
	"time"

	"github.com/teclara/sixrail/api/internal/models"
)

// SimulatePositions computes synthetic vehicle positions for all active trips at time now.
// For each trip, it finds the current segment between two stops and linearly interpolates
// the position. Trips not yet started or already finished are skipped.
func SimulatePositions(now time.Time, static *StaticStore) []models.VehiclePosition {
	loc, err := time.LoadLocation("America/Toronto")
	if err != nil {
		loc = time.UTC
	}
	nowLocal := now.In(loc)
	midnight := truncateToDay(nowLocal)
	nowOffset := nowLocal.Sub(midnight)

	trips := static.ActiveSimTrips(now)
	positions := make([]models.VehiclePosition, 0, len(trips))

	for _, trip := range trips {
		pos, ok := interpolatePosition(trip, nowOffset)
		if !ok {
			continue
		}

		route, _ := static.GetRoute(trip.RouteID)
		positions = append(positions, models.VehiclePosition{
			VehicleID:  trip.TripID,
			TripID:     trip.TripID,
			RouteID:    trip.RouteID,
			RouteName:  route.LongName,
			RouteColor: route.Color,
			Lat:        pos.lat,
			Lon:        pos.lon,
			Bearing:    pos.bearing,
			Timestamp:  now.Unix(),
		})
	}

	return positions
}

type interpResult struct {
	lat     float64
	lon     float64
	bearing float32
}

func interpolatePosition(trip SimTrip, nowOffset time.Duration) (interpResult, bool) {
	stops := trip.Stops

	if nowOffset < stops[0].DepartureTime {
		return interpResult{}, false
	}
	if nowOffset >= stops[len(stops)-1].ArrivalTime {
		return interpResult{}, false
	}

	for i := 0; i < len(stops)-1; i++ {
		a := stops[i]
		b := stops[i+1]
		if nowOffset < a.DepartureTime || nowOffset >= b.ArrivalTime {
			continue
		}
		segDur := b.ArrivalTime - a.DepartureTime
		if segDur <= 0 {
			continue
		}
		t := float64(nowOffset-a.DepartureTime) / float64(segDur)
		t = clamp(t, 0, 1)

		lat := a.Lat + t*(b.Lat-a.Lat)
		lon := a.Lon + t*(b.Lon-a.Lon)
		bearing := bearingDeg(a.Lat, a.Lon, b.Lat, b.Lon)

		return interpResult{lat: lat, lon: lon, bearing: float32(bearing)}, true
	}

	return interpResult{}, false
}

func bearingDeg(lat1, lon1, lat2, lon2 float64) float64 {
	dLon := (lon2 - lon1) * math.Pi / 180
	lat1R := lat1 * math.Pi / 180
	lat2R := lat2 * math.Pi / 180
	y := math.Sin(dLon) * math.Cos(lat2R)
	x := math.Cos(lat1R)*math.Sin(lat2R) - math.Sin(lat1R)*math.Cos(lat2R)*math.Cos(dLon)
	deg := math.Atan2(y, x) * 180 / math.Pi
	return math.Mod(deg+360, 360)
}

func clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

// StartSimulatedPositionPoller launches a background goroutine that computes
// synthetic vehicle positions from the GTFS static schedule every interval.
// Use this when no Metrolinx API key is available.
func StartSimulatedPositionPoller(ctx context.Context, static *StaticStore, cache *RealtimeCache, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		computeAndCachePositions(static, cache)

		for {
			select {
			case <-ctx.Done():
				slog.Info("simulated position poller stopped")
				return
			case <-ticker.C:
				computeAndCachePositions(static, cache)
			}
		}
	}()
}

func computeAndCachePositions(static *StaticStore, cache *RealtimeCache) {
	positions := SimulatePositions(time.Now(), static)
	cache.SetPositions(positions)
	slog.Info("simulated positions updated", "count", len(positions))
}
