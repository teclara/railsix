package gtfs

import (
	"context"
	"log/slog"
	"math"
	"time"

	"github.com/teclara/sixrail/api/internal/models"
)

// SimulatePositions computes synthetic vehicle positions for all active trips at time now.
// For each trip, it finds the current segment between two stops and interpolates
// the position along the trip's shape geometry (or straight-line if no shape).
func SimulatePositions(now time.Time, static *StaticStore) []models.VehiclePosition {
	loc, err := time.LoadLocation("America/Toronto")
	if err != nil {
		slog.Warn("failed to load America/Toronto timezone, falling back to UTC", "error", err)
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
			RouteType:  route.Type,
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
	if len(stops) < 2 {
		return interpResult{}, false
	}

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

		// Use shape geometry if available
		if len(trip.Shape) > 0 && len(trip.StopSnap) == len(stops) {
			return interpolateAlongShape(trip.Shape, trip.StopSnap[i], trip.StopSnap[i+1], t), true
		}

		// Fallback: straight-line between stops
		lat := a.Lat + t*(b.Lat-a.Lat)
		lon := a.Lon + t*(b.Lon-a.Lon)
		bearing := bearingDeg(a.Lat, a.Lon, b.Lat, b.Lon)
		return interpResult{lat: lat, lon: lon, bearing: float32(bearing)}, true
	}

	return interpResult{}, false
}

// interpolateAlongShape walks along shape points from index startIdx to endIdx,
// placing the position at fraction t (0..1) of the cumulative distance.
func interpolateAlongShape(shape []ShapePoint, startIdx, endIdx int, t float64) interpResult {
	// Ensure valid range
	if startIdx >= endIdx {
		endIdx = startIdx + 1
	}
	if endIdx >= len(shape) {
		endIdx = len(shape) - 1
	}
	if startIdx < 0 {
		startIdx = 0
	}
	if startIdx >= endIdx {
		// Degenerate: just return the point
		p := shape[startIdx]
		return interpResult{lat: p.Lat, lon: p.Lon}
	}

	// Compute cumulative distances along the shape segment
	n := endIdx - startIdx
	cumDist := make([]float64, n+1)
	cumDist[0] = 0
	for i := 1; i <= n; i++ {
		p0 := shape[startIdx+i-1]
		p1 := shape[startIdx+i]
		cumDist[i] = cumDist[i-1] + haversine(p0.Lat, p0.Lon, p1.Lat, p1.Lon)
	}

	totalDist := cumDist[n]
	if totalDist == 0 {
		p := shape[startIdx]
		return interpResult{lat: p.Lat, lon: p.Lon}
	}

	targetDist := t * totalDist

	// Find the segment containing targetDist
	for i := 1; i <= n; i++ {
		if cumDist[i] >= targetDist {
			segLen := cumDist[i] - cumDist[i-1]
			if segLen == 0 {
				p := shape[startIdx+i]
				return interpResult{lat: p.Lat, lon: p.Lon}
			}
			frac := (targetDist - cumDist[i-1]) / segLen
			p0 := shape[startIdx+i-1]
			p1 := shape[startIdx+i]
			lat := p0.Lat + frac*(p1.Lat-p0.Lat)
			lon := p0.Lon + frac*(p1.Lon-p0.Lon)
			bearing := bearingDeg(p0.Lat, p0.Lon, p1.Lat, p1.Lon)
			return interpResult{lat: lat, lon: lon, bearing: float32(bearing)}
		}
	}

	// Shouldn't reach here, but return end point
	p := shape[endIdx]
	bearing := bearingDeg(shape[endIdx-1].Lat, shape[endIdx-1].Lon, p.Lat, p.Lon)
	return interpResult{lat: p.Lat, lon: p.Lon, bearing: float32(bearing)}
}

func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371000 // Earth radius in meters
	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180
	lat1R := lat1 * math.Pi / 180
	lat2R := lat2 * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1R)*math.Cos(lat2R)*math.Sin(dLon/2)*math.Sin(dLon/2)
	return R * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
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
