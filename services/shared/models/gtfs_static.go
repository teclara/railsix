package models

// ScheduledDeparture is a single stop-time entry in the schedule index.
type ScheduledDeparture struct {
	TripID        string `json:"tripId"`
	RouteID       string `json:"routeId"`
	ServiceID     string `json:"serviceId"`
	Headsign      string `json:"headsign"`
	DepartureTime int64  `json:"departureTime"` // nanoseconds from midnight (local time)
}

// TripStop is one stop in a trip's ordered sequence.
type TripStop struct {
	StopID        string `json:"stopId"`
	StopCode      string `json:"stopCode"`
	ArrivalTime   int64  `json:"arrivalTime"`   // nanoseconds from midnight of service day
	DepartureTime int64  `json:"departureTime"` // nanoseconds from midnight of service day
}

// TripInfo holds a trip's identity and full stop sequence for departure enrichment.
type TripInfo struct {
	TripID    string     `json:"tripId"`
	RouteID   string     `json:"routeId"`
	ServiceID string     `json:"serviceId"`
	Stops     []TripStop `json:"stops"`
}

// ArrivalResult is the JSON-serializable result of an arrival time query.
type ArrivalResult struct {
	Duration int64 `json:"duration"` // nanoseconds
	OK       bool  `json:"ok"`
}

// ScheduleCandidate is a pre-filtered departure candidate returned by gtfs-static.
type ScheduleCandidate struct {
	TripID         string   `json:"tripId"`
	TripNumber     string   `json:"tripNumber"`
	RouteShortName string   `json:"routeShortName"`
	RouteLongName  string   `json:"routeLongName"`
	RouteColor     string   `json:"routeColor"`
	RouteType      int      `json:"routeType"`
	Headsign       string   `json:"headsign"`
	ScheduledTime  string   `json:"scheduledTime"` // "HH:MM"
	Platform       string   `json:"platform"`
	Stops          []string `json:"stops"`        // remaining stop names after departure
	LastStopCode   string   `json:"lastStopCode"` // public stop code of final destination
	IsExpress      bool     `json:"isExpress"`
	StopID         string   `json:"stopId"`
	StopCode       string   `json:"stopCode"`
	DepartureNano  int64    `json:"departureNano"` // nanoseconds from midnight of service day
	ServiceDay     string   `json:"serviceDay"`    // "YYYY-MM-DD"
}
