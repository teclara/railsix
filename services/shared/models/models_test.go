package models

import (
	"encoding/json"
	"testing"
)

func TestDepartureRoundTrip(t *testing.T) {
	original := Departure{
		Line:          "LW",
		LineName:      "Lakeshore West",
		Destination:   "Union Station",
		ScheduledTime: "08:30",
		ActualTime:    "08:32",
		ArrivalTime:   "09:15",
		Status:        "Delayed +2m",
		Platform:      "3",
		RouteColor:    "00853F",
		DelayMinutes:  2,
		Stops:         []string{"Oakville", "Clarkson", "Port Credit"},
		Cars:          "12",
		IsInMotion:    true,
		IsCancelled:   false,
		IsExpress:     true,
		RouteType:     2,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("failed to marshal Departure: %v", err)
	}

	var decoded Departure
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal Departure: %v", err)
	}

	if decoded.Line != original.Line {
		t.Errorf("Line: got %q, want %q", decoded.Line, original.Line)
	}
	if decoded.LineName != original.LineName {
		t.Errorf("LineName: got %q, want %q", decoded.LineName, original.LineName)
	}
	if decoded.Destination != original.Destination {
		t.Errorf("Destination: got %q, want %q", decoded.Destination, original.Destination)
	}
	if decoded.ScheduledTime != original.ScheduledTime {
		t.Errorf("ScheduledTime: got %q, want %q", decoded.ScheduledTime, original.ScheduledTime)
	}
	if decoded.ActualTime != original.ActualTime {
		t.Errorf("ActualTime: got %q, want %q", decoded.ActualTime, original.ActualTime)
	}
	if decoded.ArrivalTime != original.ArrivalTime {
		t.Errorf("ArrivalTime: got %q, want %q", decoded.ArrivalTime, original.ArrivalTime)
	}
	if decoded.Status != original.Status {
		t.Errorf("Status: got %q, want %q", decoded.Status, original.Status)
	}
	if decoded.Platform != original.Platform {
		t.Errorf("Platform: got %q, want %q", decoded.Platform, original.Platform)
	}
	if decoded.RouteColor != original.RouteColor {
		t.Errorf("RouteColor: got %q, want %q", decoded.RouteColor, original.RouteColor)
	}
	if decoded.DelayMinutes != original.DelayMinutes {
		t.Errorf("DelayMinutes: got %d, want %d", decoded.DelayMinutes, original.DelayMinutes)
	}
	if len(decoded.Stops) != len(original.Stops) {
		t.Errorf("Stops length: got %d, want %d", len(decoded.Stops), len(original.Stops))
	} else {
		for i, s := range decoded.Stops {
			if s != original.Stops[i] {
				t.Errorf("Stops[%d]: got %q, want %q", i, s, original.Stops[i])
			}
		}
	}
	if decoded.Cars != original.Cars {
		t.Errorf("Cars: got %q, want %q", decoded.Cars, original.Cars)
	}
	if decoded.IsInMotion != original.IsInMotion {
		t.Errorf("IsInMotion: got %v, want %v", decoded.IsInMotion, original.IsInMotion)
	}
	if decoded.IsCancelled != original.IsCancelled {
		t.Errorf("IsCancelled: got %v, want %v", decoded.IsCancelled, original.IsCancelled)
	}
	if decoded.IsExpress != original.IsExpress {
		t.Errorf("IsExpress: got %v, want %v", decoded.IsExpress, original.IsExpress)
	}
	if decoded.RouteType != original.RouteType {
		t.Errorf("RouteType: got %d, want %d", decoded.RouteType, original.RouteType)
	}
}

func TestDepartureOmitEmpty(t *testing.T) {
	minimal := Departure{
		Line:          "LW",
		Destination:   "Union",
		ScheduledTime: "08:30",
		Status:        "On Time",
	}

	data, err := json.Marshal(minimal)
	if err != nil {
		t.Fatalf("failed to marshal minimal Departure: %v", err)
	}

	jsonStr := string(data)
	// These omitempty fields should not appear
	for _, field := range []string{"lineName", "actualTime", "arrivalTime", "platform", "routeColor", "stops", "cars"} {
		if contains(jsonStr, `"`+field+`"`) {
			t.Errorf("expected omitempty field %q to be absent, but found in JSON: %s", field, jsonStr)
		}
	}
}

func TestAlertRoundTrip(t *testing.T) {
	original := Alert{
		ID:          "alert-123",
		Effect:      "DETOUR",
		Headline:    "Service disruption on Lakeshore West",
		Description: "Due to signal issues, trains are delayed.",
		URL:         "https://gotransit.com/alerts/123",
		RouteIDs:    []string{"01", "02"},
		RouteNames:  []string{"Lakeshore West", "Lakeshore East"},
		StartTime:   1700000000,
		EndTime:     1700100000,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("failed to marshal Alert: %v", err)
	}

	var decoded Alert
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal Alert: %v", err)
	}

	if decoded.ID != original.ID {
		t.Errorf("ID: got %q, want %q", decoded.ID, original.ID)
	}
	if decoded.Effect != original.Effect {
		t.Errorf("Effect: got %q, want %q", decoded.Effect, original.Effect)
	}
	if decoded.Headline != original.Headline {
		t.Errorf("Headline: got %q, want %q", decoded.Headline, original.Headline)
	}
	if decoded.Description != original.Description {
		t.Errorf("Description: got %q, want %q", decoded.Description, original.Description)
	}
	if decoded.URL != original.URL {
		t.Errorf("URL: got %q, want %q", decoded.URL, original.URL)
	}
	if len(decoded.RouteIDs) != len(original.RouteIDs) {
		t.Errorf("RouteIDs length: got %d, want %d", len(decoded.RouteIDs), len(original.RouteIDs))
	} else {
		for i, id := range decoded.RouteIDs {
			if id != original.RouteIDs[i] {
				t.Errorf("RouteIDs[%d]: got %q, want %q", i, id, original.RouteIDs[i])
			}
		}
	}
	if len(decoded.RouteNames) != len(original.RouteNames) {
		t.Errorf("RouteNames length: got %d, want %d", len(decoded.RouteNames), len(original.RouteNames))
	} else {
		for i, name := range decoded.RouteNames {
			if name != original.RouteNames[i] {
				t.Errorf("RouteNames[%d]: got %q, want %q", i, name, original.RouteNames[i])
			}
		}
	}
	if decoded.StartTime != original.StartTime {
		t.Errorf("StartTime: got %d, want %d", decoded.StartTime, original.StartTime)
	}
	if decoded.EndTime != original.EndTime {
		t.Errorf("EndTime: got %d, want %d", decoded.EndTime, original.EndTime)
	}
}

func TestAlertOmitEmpty(t *testing.T) {
	minimal := Alert{
		ID:          "alert-456",
		Effect:      "NO_SERVICE",
		Headline:    "No trains",
		Description: "All service suspended.",
	}

	data, err := json.Marshal(minimal)
	if err != nil {
		t.Fatalf("failed to marshal minimal Alert: %v", err)
	}

	jsonStr := string(data)
	for _, field := range []string{"url", "routeIds", "routeNames", "startTime", "endTime"} {
		if contains(jsonStr, `"`+field+`"`) {
			t.Errorf("expected omitempty field %q to be absent, but found in JSON: %s", field, jsonStr)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
