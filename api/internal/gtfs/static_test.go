package gtfs_test

import (
	"archive/zip"
	"bytes"
	"testing"
	"time"

	gtfsstore "github.com/teclara/sixrail/api/internal/gtfs"
)

// buildTestZip creates a minimal GTFS zip with stops.txt, routes.txt,
// agency.txt, calendar.txt, and trips.txt (required by parser).
func buildTestZip(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)

	files := map[string]string{
		"agency.txt":   "agency_id,agency_name,agency_url,agency_timezone\nMX,Metrolinx,https://metrolinx.com,America/Toronto\n",
		"calendar.txt": "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\nWD,1,1,1,1,1,0,0,20260101,20261231\n",
		"routes.txt":   "route_id,agency_id,route_short_name,route_long_name,route_type,route_color,route_text_color\n01,MX,LW,Lakeshore West,2,098137,FFFFFF\n09,MX,LE,Lakeshore East,2,098137,FFFFFF\n",
		"stops.txt":    "stop_id,stop_code,stop_name,stop_lat,stop_lon,location_type,parent_station\nUN,UN,Union Station,43.6453,-79.3806,1,\nUNp1,UNp1,Union Station Platform 1,43.6454,-79.3807,0,UN\n",
		"trips.txt":      "route_id,service_id,trip_id,direction_id\n01,WD,T001,0\n",
		"stop_times.txt": "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT001,08:00:00,08:00:00,UN,1\n",
	}

	for name, content := range files {
		f, err := w.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		f.Write([]byte(content))
	}
	w.Close()
	return buf.Bytes()
}

func TestStaticStore_LoadFromBytes(t *testing.T) {
	zipData := buildTestZip(t)
	store, err := gtfsstore.NewStaticStore(zipData)
	if err != nil {
		t.Fatal(err)
	}

	stops := store.AllStops()
	if len(stops) == 0 {
		t.Fatal("expected stops")
	}

	found := false
	for _, s := range stops {
		if s.Code == "UN" {
			found = true
			if s.Name != "Union Station" {
				t.Fatalf("expected Union Station, got %s", s.Name)
			}
		}
	}
	if !found {
		t.Fatal("Union Station not found")
	}
}

func TestStaticStore_GetRoute(t *testing.T) {
	zipData := buildTestZip(t)
	store, err := gtfsstore.NewStaticStore(zipData)
	if err != nil {
		t.Fatal(err)
	}

	r, ok := store.GetRoute("01")
	if !ok {
		t.Fatal("expected route 01")
	}
	if r.ShortName != "LW" {
		t.Fatalf("expected LW, got %s", r.ShortName)
	}
	if r.Color != "098137" {
		t.Fatalf("expected 098137, got %s", r.Color)
	}
}

func TestStaticStore_GetRoute_NotFound(t *testing.T) {
	zipData := buildTestZip(t)
	store, err := gtfsstore.NewStaticStore(zipData)
	if err != nil {
		t.Fatal(err)
	}

	_, ok := store.GetRoute("99")
	if ok {
		t.Fatal("expected route 99 to not be found")
	}
}

func buildSimTestZip(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	files := map[string]string{
		"agency.txt":   "agency_id,agency_name,agency_url,agency_timezone\nMX,Metrolinx,https://metrolinx.com,America/Toronto\n",
		"calendar.txt": "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\nWD,1,1,1,1,1,0,0,20260101,20261231\n",
		"routes.txt":   "route_id,agency_id,route_short_name,route_long_name,route_type,route_color,route_text_color\n01,MX,LW,Lakeshore West,2,098137,FFFFFF\n",
		"stops.txt":    "stop_id,stop_code,stop_name,stop_lat,stop_lon,location_type,parent_station\nUN,UN,Union Station,43.6453,-79.3806,1,\nMI,MI,Mimico,43.6200,-79.4900,1,\n",
		"trips.txt":      "route_id,service_id,trip_id,direction_id\n01,WD,T001,0\n",
		"stop_times.txt": "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT001,08:00:00,08:00:00,UN,1\nT001,08:20:00,08:20:00,MI,2\n",
	}
	for name, content := range files {
		f, _ := w.Create(name)
		f.Write([]byte(content))
	}
	w.Close()
	return buf.Bytes()
}

func TestStaticStore_ActiveSimTrips_WeekdayMatch(t *testing.T) {
	store, err := gtfsstore.NewStaticStore(buildSimTestZip(t))
	if err != nil {
		t.Fatal(err)
	}
	// 2026-03-03 is a Tuesday — should match WD (weekday) service
	tuesday := time.Date(2026, 3, 3, 9, 0, 0, 0, time.UTC)
	trips := store.ActiveSimTrips(tuesday)
	if len(trips) != 1 {
		t.Fatalf("expected 1 active trip, got %d", len(trips))
	}
	if trips[0].TripID != "T001" {
		t.Fatalf("expected T001, got %s", trips[0].TripID)
	}
	if len(trips[0].Stops) != 2 {
		t.Fatalf("expected 2 stops, got %d", len(trips[0].Stops))
	}
}

func TestStaticStore_ActiveSimTrips_WeekendNoMatch(t *testing.T) {
	store, err := gtfsstore.NewStaticStore(buildSimTestZip(t))
	if err != nil {
		t.Fatal(err)
	}
	// 2026-03-07 is a Saturday — WD service is not active
	saturday := time.Date(2026, 3, 7, 9, 0, 0, 0, time.UTC)
	trips := store.ActiveSimTrips(saturday)
	if len(trips) != 0 {
		t.Fatalf("expected 0 active trips on weekend, got %d", len(trips))
	}
}
