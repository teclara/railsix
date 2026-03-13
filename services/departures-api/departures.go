package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/teclara/railsix/shared/gtfsrt"
	"github.com/teclara/railsix/shared/models"
)

var torontoLoc *time.Location

func init() {
	var err error
	torontoLoc, err = time.LoadLocation("America/Toronto")
	if err != nil {
		panic("failed to load America/Toronto timezone: " + err.Error())
	}
}

// GetDepartures returns upcoming departures for a stop code, merging static schedule
// with real-time trip updates. If destCode is non-empty, ArrivalTime is populated
// for each departure.
func GetDepartures(ctx context.Context, stopCode, destCode string, now time.Time, sc *StaticClient, rc *RedisClient) ([]models.Departure, error) {
	nowLocal := now.In(torontoLoc)

	candidates, err := sc.GetSchedule(stopCode, nowLocal)
	if err != nil {
		return nil, fmt.Errorf("get schedule: %w", err)
	}
	if len(candidates) == 0 {
		return []models.Departure{}, nil
	}

	var destStopIDs []string
	if destCode != "" {
		destStopIDs, err = sc.StopIDsForCode(destCode)
		if err != nil {
			return nil, fmt.Errorf("get destination stop IDs: %w", err)
		}
	}

	if err := rc.RequireFresh(ctx, keyTripUpdatesUpdatedAt, "trip updates", realtimeFreshnessThreshold); err != nil {
		return nil, err
	}
	if err := rc.RequireFresh(ctx, keyServiceGlanceUpdatedAt, "service glance", realtimeFreshnessThreshold); err != nil {
		return nil, err
	}
	if err := rc.RequireFresh(ctx, keyExceptionsUpdatedAt, "exceptions", realtimeFreshnessThreshold); err != nil {
		return nil, err
	}

	tripUpdates, err := rc.GetAllTripUpdates(ctx)
	if err != nil {
		return nil, err
	}
	glanceAll, err := rc.GetAllServiceGlanceMap(ctx)
	if err != nil {
		return nil, err
	}
	exceptions, err := rc.GetAllExceptions(ctx)
	if err != nil {
		return nil, err
	}

	unionByTrip := map[string]models.UnionDeparture{}
	if err := rc.RequireFresh(ctx, keyUnionDeparturesUpdatedAt, "union departures", realtimeFreshnessThreshold); err == nil {
		unionDeps, readErr := rc.GetUnionDepartures(ctx)
		if readErr != nil {
			return nil, readErr
		}
		unionByTrip = make(map[string]models.UnionDeparture, len(unionDeps))
		for _, ud := range unionDeps {
			unionByTrip[ud.TripNumber] = ud
		}
	}

	result := make([]models.Departure, 0, len(candidates))
	for i := range candidates {
		c := &candidates[i]
		serviceDay, _ := time.ParseInLocation("2006-01-02", c.ServiceDay, torontoLoc)
		scheduled := serviceDay.Add(time.Duration(c.DepartureNano))

		update, ok := findTripUpdate(tripUpdates, c.TripID)
		delay := normalizeDepartureDelay(findDelay(update, ok, c.StopID))
		adjusted := scheduled.Add(delay)
		delayMin := int(delay.Minutes())

		status := departureStatus(ok, update.ScheduleRelationship, delayMin)

		dep := models.Departure{
			Line:          c.RouteShortName,
			LineName:      c.RouteLongName,
			Destination:   c.Headsign,
			ScheduledTime: c.ScheduledTime,
			Status:        status,
			Platform:      c.Platform,
			RouteColor:    c.RouteColor,
			DelayMinutes:  delayMin,
			Stops:         c.Stops,
			LastStopCode:  c.LastStopCode,
			IsExpress:     c.IsExpress,
			RouteType:     c.RouteType,
			TripNumber:    c.TripNumber,
		}
		if delayMin != 0 {
			dep.ActualTime = formatTime(adjusted)
		}
		if len(destStopIDs) > 0 {
			arr, arrivalErr := sc.ArrivalTimeAtStop(c.TripID, destStopIDs, []string{c.StopID})
			if arrivalErr != nil {
				return nil, fmt.Errorf("get arrival time for %s: %w", c.TripID, arrivalErr)
			}
			if !arr.OK {
				continue
			}
			dep.ArrivalTime = formatTime(serviceDay.Add(time.Duration(arr.Duration)))
		}

		if sg, ok := glanceAll[c.TripNumber]; ok {
			dep.Cars = sg.Cars
			dep.IsInMotion = sg.IsInMotion
			if dep.Status != "Cancelled" && delay == 0 && sg.DelaySeconds != 0 {
				sgDelay := normalizeDepartureDelay(time.Duration(sg.DelaySeconds) * time.Second)
				adjusted = scheduled.Add(sgDelay)
				delayMin = int(sgDelay.Minutes())
				dep.DelayMinutes = delayMin
				if delayMin != 0 {
					dep.Status = departureStatus(true, "", delayMin)
					dep.ActualTime = formatTime(adjusted)
				}
			}
		}

		if ud, ok := unionByTrip[c.TripNumber]; ok {
			isUnion := strings.EqualFold(stopCode, "UN")
			if isUnion && ud.Platform != "" && dep.Platform == "" {
				dep.Platform = ud.Platform
			}
			if ud.Info != "" {
				dep.Status = ud.Info
			}
			if strings.Contains(strings.ToUpper(ud.Info), "CANCEL") {
				dep.IsCancelled = true
				dep.Status = "Cancelled"
			}
		}

		if cancelledStops, ok := exceptions[c.TripNumber]; ok && isStopCancelled(cancelledStops, c.StopCode) {
			dep.IsCancelled = true
			dep.Status = "Cancelled"
		}

		result = append(result, dep)
	}

	return result, nil
}

// findTripUpdate looks up a trip update by full trip ID first, then by trip number.
func findTripUpdate(tripUpdates map[string]gtfsrt.RawTripUpdate, tripID string) (gtfsrt.RawTripUpdate, bool) {
	if update, ok := tripUpdates[tripID]; ok {
		return update, true
	}
	update, ok := tripUpdates[models.ExtractTripNumber(tripID)]
	return update, ok
}

// findDelay returns the departure delay for a trip at a given stop.
func findDelay(update gtfsrt.RawTripUpdate, ok bool, stopID string) time.Duration {
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
	return propagated
}

func isStopCancelled(cancelledStops []string, stopCode string) bool {
	if len(cancelledStops) == 0 {
		return true
	}
	for _, s := range cancelledStops {
		if s == stopCode {
			return true
		}
	}
	return false
}

func departureStatus(hasUpdate bool, scheduleRelationship string, delayMin int) string {
	if hasUpdate && scheduleRelationship == "CANCELED" {
		return "Cancelled"
	}
	if delayMin >= 1 {
		return fmt.Sprintf("Delayed +%dm", delayMin)
	}
	return "On Time"
}

func normalizeDepartureDelay(delay time.Duration) time.Duration {
	if delay < 0 {
		return -delay
	}
	return delay
}

// formatTime returns "HH:MM" in local time.
func formatTime(t time.Time) string {
	return fmt.Sprintf("%02d:%02d", t.Hour(), t.Minute())
}
