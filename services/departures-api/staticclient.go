package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/teclara/railsix/shared/models"
)

// ScheduledDeparture mirrors gtfs-static's store.ScheduledDeparture.
type ScheduledDeparture struct {
	TripID        string `json:"tripId"`
	RouteID       string `json:"routeId"`
	ServiceID     string `json:"serviceId"`
	Headsign      string `json:"headsign"`
	DepartureTime int64  `json:"departureTime"` // nanoseconds from midnight
}

// TripStop mirrors gtfs-static's store.TripStop.
type TripStop struct {
	StopID        string `json:"stopId"`
	ArrivalTime   int64  `json:"arrivalTime"`   // nanoseconds from midnight
	DepartureTime int64  `json:"departureTime"` // nanoseconds from midnight
}

// TripInfo mirrors gtfs-static's store.TripInfo.
type TripInfo struct {
	TripID    string     `json:"tripId"`
	RouteID   string     `json:"routeId"`
	ServiceID string     `json:"serviceId"`
	Stops     []TripStop `json:"stops"`
}

// ArrivalResult mirrors gtfs-static's store.ArrivalResult.
type ArrivalResult struct {
	Duration int64 `json:"duration"` // nanoseconds
	OK       bool  `json:"ok"`
}

// StaticClient is an HTTP client for the gtfs-static microservice.
type StaticClient struct {
	baseURL string
	client  *http.Client
}

// NewStaticClient creates a StaticClient pointing at the given base URL.
func NewStaticClient(baseURL string) *StaticClient {
	return &StaticClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (sc *StaticClient) get(path string) ([]byte, error) {
	resp, err := sc.client.Get(sc.baseURL + path)
	if err != nil {
		return nil, fmt.Errorf("static client GET %s: %w", path, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("static client GET %s: status %d", path, resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

// StopIDsForCode returns all stop IDs for a stop code.
func (sc *StaticClient) StopIDsForCode(code string) ([]string, error) {
	data, err := sc.get("/stops/" + url.PathEscape(code) + "/ids")
	if err != nil {
		return nil, err
	}
	var ids []string
	if err := json.Unmarshal(data, &ids); err != nil {
		return nil, fmt.Errorf("decode stop ids: %w", err)
	}
	return ids, nil
}

// DeparturesForStop returns scheduled departures for a stop ID.
func (sc *StaticClient) DeparturesForStop(stopID string) ([]ScheduledDeparture, error) {
	data, err := sc.get("/departures/" + url.PathEscape(stopID))
	if err != nil {
		return nil, err
	}
	var deps []ScheduledDeparture
	if err := json.Unmarshal(data, &deps); err != nil {
		return nil, fmt.Errorf("decode departures: %w", err)
	}
	return deps, nil
}

// IsLastStop returns true if any of the given stop IDs is the final stop of the trip.
func (sc *StaticClient) IsLastStop(tripID string, stopIDs []string) (bool, error) {
	params := url.Values{}
	for _, id := range stopIDs {
		params.Add("stopID", id)
	}
	data, err := sc.get("/trips/" + url.PathEscape(tripID) + "/is-last-stop?" + params.Encode())
	if err != nil {
		return false, err
	}
	var result struct {
		IsLastStop bool `json:"isLastStop"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return false, fmt.Errorf("decode is-last-stop: %w", err)
	}
	return result.IsLastStop, nil
}

// IsServiceActive returns whether a service is active on a given date.
func (sc *StaticClient) IsServiceActive(serviceID string, date time.Time) (bool, error) {
	dateStr := date.Format("2006-01-02")
	data, err := sc.get("/services/" + url.PathEscape(serviceID) + "/active?date=" + dateStr)
	if err != nil {
		return false, err
	}
	var result struct {
		Active bool `json:"active"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return false, fmt.Errorf("decode service active: %w", err)
	}
	return result.Active, nil
}

// GetRoute returns route info for a route ID.
func (sc *StaticClient) GetRoute(routeID string) (models.Route, bool) {
	data, err := sc.get("/routes/" + url.PathEscape(routeID))
	if err != nil {
		return models.Route{}, false
	}
	var route models.Route
	if err := json.Unmarshal(data, &route); err != nil {
		return models.Route{}, false
	}
	return route, true
}

// GetStopName returns the name for a stop ID.
func (sc *StaticClient) GetStopName(stopID string) (string, error) {
	data, err := sc.get("/stop-name/" + url.PathEscape(stopID))
	if err != nil {
		return "", err
	}
	var result struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("decode stop name: %w", err)
	}
	return result.Name, nil
}

// RemainingStopNames returns stop names after the departure stop in a trip.
func (sc *StaticClient) RemainingStopNames(tripID string, stopIDs []string) ([]string, error) {
	params := url.Values{}
	for _, id := range stopIDs {
		params.Add("stopID", id)
	}
	data, err := sc.get("/trips/" + url.PathEscape(tripID) + "/remaining-stops?" + params.Encode())
	if err != nil {
		return nil, err
	}
	var names []string
	if err := json.Unmarshal(data, &names); err != nil {
		return nil, fmt.Errorf("decode remaining stops: %w", err)
	}
	return names, nil
}

// IsExpress returns whether a trip is express (skips stops).
func (sc *StaticClient) IsExpress(tripID string) (bool, error) {
	data, err := sc.get("/trips/" + url.PathEscape(tripID) + "/is-express")
	if err != nil {
		return false, err
	}
	var result struct {
		IsExpress bool `json:"isExpress"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return false, fmt.Errorf("decode is-express: %w", err)
	}
	return result.IsExpress, nil
}

// ArrivalTimeAtStop returns the arrival duration at a destination stop.
func (sc *StaticClient) ArrivalTimeAtStop(tripID string, destIDs, originIDs []string) (ArrivalResult, error) {
	params := url.Values{}
	for _, id := range destIDs {
		params.Add("destID", id)
	}
	for _, id := range originIDs {
		params.Add("originID", id)
	}
	data, err := sc.get("/trips/" + url.PathEscape(tripID) + "/arrival?" + params.Encode())
	if err != nil {
		return ArrivalResult{}, err
	}
	var result ArrivalResult
	if err := json.Unmarshal(data, &result); err != nil {
		return ArrivalResult{}, fmt.Errorf("decode arrival: %w", err)
	}
	return result, nil
}
