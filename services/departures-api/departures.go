package main

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/teclara/railsix/shared/gtfsrt"
	"github.com/teclara/railsix/shared/models"
)

const (
	torontoTZ      = "America/Toronto"
	maxDepartures  = 20
	lookAheadHours = 3
)

// GetDepartures returns upcoming departures for a stop code, merging static schedule
// with real-time trip updates. Falls back gracefully if updates are unavailable.
// If destCode is non-empty, ArrivalTime is populated for each departure.
func GetDepartures(ctx context.Context, stopCode, destCode string, now time.Time, sc *StaticClient, rc *RedisClient) []models.Departure {
	loc, err := time.LoadLocation(torontoTZ)
	if err != nil {
		panic("failed to load America/Toronto timezone: " + err.Error())
	}
	nowLocal := now.In(loc)

	stopIDs, err := sc.StopIDsForCode(stopCode)
	if err != nil || len(stopIDs) == 0 {
		if err != nil {
			slog.Warn("failed to get stop IDs", "stopCode", stopCode, "error", err)
		}
		return []models.Departure{}
	}

	var destStopIDs []string
	if destCode != "" {
		destStopIDs, err = sc.StopIDsForCode(destCode)
		if err != nil {
			slog.Warn("failed to get dest stop IDs", "destCode", destCode, "error", err)
		}
	}

	// Determine active service IDs for today (and yesterday for past-midnight trips).
	today := truncateToDay(nowLocal)
	yesterday := today.Add(-24 * time.Hour)

	type candidate struct {
		dep        ScheduledDeparture
		stopID     string
		serviceDay time.Time
		adjusted   time.Time
	}

	var candidates []candidate

	for _, stopID := range stopIDs {
		departures, err := sc.DeparturesForStop(stopID)
		if err != nil {
			slog.Warn("failed to get departures for stop", "stopID", stopID, "error", err)
			continue
		}
		for _, dep := range departures {
			// Skip trips where this stop is the final stop.
			isLast, err := sc.IsLastStop(dep.TripID, stopIDs)
			if err != nil {
				slog.Debug("failed to check is-last-stop", "tripID", dep.TripID, "error", err)
				continue
			}
			if isLast {
				continue
			}
			// Try both today and yesterday (for past-midnight services).
			for _, serviceDay := range []time.Time{today, yesterday} {
				active, err := sc.IsServiceActive(dep.ServiceID, serviceDay)
				if err != nil {
					slog.Debug("failed to check service active", "serviceID", dep.ServiceID, "error", err)
					continue
				}
				if !active {
					continue
				}
				// Compute wall-clock scheduled departure time.
				// DepartureTime is nanoseconds from midnight of the service day.
				scheduled := serviceDay.Add(time.Duration(dep.DepartureTime))

				// Apply real-time delay if available.
				delay := findDelay(ctx, dep.TripID, stopID, rc)
				adjusted := scheduled.Add(delay)

				// Only include if within the look-ahead window and not in the past.
				if adjusted.Before(nowLocal) {
					continue
				}
				if adjusted.After(nowLocal.Add(lookAheadHours * time.Hour)) {
					continue
				}

				candidates = append(candidates, candidate{dep, stopID, serviceDay, adjusted})
				break // matched a service day — no need to check the other
			}
		}
	}

	// Sort by adjusted departure time.
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].adjusted.Before(candidates[j].adjusted)
	})

	// Deduplicate by trip number AND by departure time + line.
	seenTrip := make(map[string]bool)
	seenTimeLine := make(map[string]bool)
	result := make([]models.Departure, 0, maxDepartures)
	for _, c := range candidates {
		tripNum := extractTripNumber(c.dep.TripID)
		if seenTrip[tripNum] {
			continue
		}
		seenTrip[tripNum] = true

		// Also dedup by scheduled time + route + stop.
		timeLineKey := formatTime(c.serviceDay.Add(time.Duration(c.dep.DepartureTime))) + "|" + c.dep.RouteID + "|" + c.stopID
		if seenTimeLine[timeLineKey] {
			continue
		}
		seenTimeLine[timeLineKey] = true

		route, _ := sc.GetRoute(c.dep.RouteID)
		delay := c.adjusted.Sub(c.serviceDay.Add(time.Duration(c.dep.DepartureTime)))
		delayMin := int(delay.Minutes())

		status := "On Time"
		if update, ok := findTripUpdate(ctx, c.dep.TripID, rc); ok {
			if update.ScheduleRelationship == "CANCELED" {
				status = "Cancelled"
			} else if delayMin >= 1 {
				status = fmt.Sprintf("Delayed +%dm", delayMin)
			}
		}

		stopName, _ := sc.GetStopName(c.stopID)
		remainingStops, _ := sc.RemainingStopNames(c.dep.TripID, stopIDs)
		isExpress, _ := sc.IsExpress(c.dep.TripID)

		dep := models.Departure{
			Line:          route.ShortName,
			LineName:      route.LongName,
			Destination:   c.dep.Headsign,
			ScheduledTime: formatTime(c.serviceDay.Add(time.Duration(c.dep.DepartureTime))),
			Status:        status,
			Platform:      extractPlatform(stopName),
			RouteColor:    route.Color,
			DelayMinutes:  delayMin,
			Stops:         remainingStops,
			IsExpress:     isExpress,
			RouteType:     route.Type,
		}
		if delayMin > 0 {
			dep.ActualTime = formatTime(c.adjusted)
		}
		if len(destStopIDs) > 0 {
			arr, err := sc.ArrivalTimeAtStop(c.dep.TripID, destStopIDs, stopIDs)
			if err != nil {
				slog.Debug("failed to get arrival time", "tripID", c.dep.TripID, "error", err)
				continue
			}
			if !arr.OK {
				continue // skip trips that don't stop at the destination
			}
			dep.ArrivalTime = formatTime(c.serviceDay.Add(time.Duration(arr.Duration)))
		}

		// Enrich with service glance data.
		tripNumber := extractTripNumber(c.dep.TripID)
		if sg, ok := rc.GetServiceGlanceEntry(ctx, tripNumber); ok {
			dep.Cars = sg.Cars
			dep.IsInMotion = sg.IsInMotion
		}

		// Enrich with Union departures board info.
		if ud, ok := rc.GetUnionDepartureByTrip(ctx, tripNumber); ok {
			isUnion := strings.EqualFold(stopCode, "UN")
			if isUnion && ud.Platform != "" && dep.Platform == "" {
				dep.Platform = ud.Platform
			}
			if ud.Info != "" {
				dep.Status = ud.Info
			}
		}

		// Flag cancelled trips from exceptions cache.
		if rc.IsTripCancelled(ctx, tripNumber) {
			dep.IsCancelled = true
			dep.Status = "Cancelled"
		}

		result = append(result, dep)

		if len(result) >= maxDepartures {
			break
		}
	}

	return result
}

// findTripUpdate looks up a trip update by full trip ID first, then by trip number.
func findTripUpdate(ctx context.Context, tripID string, rc *RedisClient) (gtfsrt.RawTripUpdate, bool) {
	if update, ok := rc.GetTripUpdate(ctx, tripID); ok {
		return update, true
	}
	return rc.GetTripUpdate(ctx, extractTripNumber(tripID))
}

// findDelay returns the departure delay for a trip at a given stop.
func findDelay(ctx context.Context, tripID, stopID string, rc *RedisClient) time.Duration {
	update, ok := findTripUpdate(ctx, tripID, rc)
	if !ok {
		return 0
	}
	var propagated time.Duration
	for _, stu := range update.StopTimeUpdates {
		if stu.DepartureDelay != 0 {
			propagated = stu.DepartureDelay
		}
		if stu.StopID == stopID {
			if stu.DepartureDelay != 0 {
				return stu.DepartureDelay
			}
			return propagated
		}
	}
	return 0
}

// extractTripNumber returns the Metrolinx trip number from a GTFS trip ID.
// GTFS trip IDs have the format "20260424-LW-1731"; the trip number is the last segment.
func extractTripNumber(tripID string) string {
	if idx := strings.LastIndex(tripID, "-"); idx >= 0 && idx+1 < len(tripID) {
		return tripID[idx+1:]
	}
	return tripID
}

// extractPlatform extracts the platform number from a GTFS stop name.
// e.g. "Oakville GO Platform 1" -> "1", "Union Station Platform 12" -> "12"
func extractPlatform(stopName string) string {
	const prefix = "Platform "
	if idx := strings.LastIndex(stopName, prefix); idx >= 0 {
		return stopName[idx+len(prefix):]
	}
	return ""
}

// formatTime returns "HH:MM" in local time.
func formatTime(t time.Time) string {
	return fmt.Sprintf("%02d:%02d", t.Hour(), t.Minute())
}

// truncateToDay returns midnight (00:00) of the given day in the same location.
func truncateToDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}
