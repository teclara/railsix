package gtfs_test

import (
	"archive/zip"
	"bytes"
	"math"
	"testing"
	"time"

	gtfsstore "github.com/teclara/sixrail/api/internal/gtfs"
)

func buildSimZip(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	files := map[string]string{
		"agency.txt":   "agency_id,agency_name,agency_url,agency_timezone\nMX,Metrolinx,https://metrolinx.com,America/Toronto\n",
		"calendar.txt": "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\nWD,1,1,1,1,1,0,0,20260101,20261231\n",
		"routes.txt":   "route_id,agency_id,route_short_name,route_long_name,route_type,route_color,route_text_color\n01,MX,LW,Lakeshore West,2,098137,FFFFFF\n",
		"stops.txt": "stop_id,stop_code,stop_name,stop_lat,stop_lon,location_type,parent_station\nUN,UN,Union Station,43.6453,-79.3806,1,\nMI,MI,Mimico,43.6200,-79.4900,1,\n",
		"trips.txt":  "route_id,service_id,trip_id,direction_id\n01,WD,T001,0\n",
		"stop_times.txt": "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT001,09:00:00,09:00:00,UN,1\nT001,09:20:00,09:20:00,MI,2\n",
	}
	for name, content := range files {
		f, _ := w.Create(name)
		f.Write([]byte(content))
	}
	w.Close()
	return buf.Bytes()
}

func TestSimulatePositions_Midpoint(t *testing.T) {
	store, err := gtfsstore.NewStaticStore(buildSimZip(t))
	if err != nil {
		t.Fatal(err)
	}
	// Tuesday 2026-03-03, 09:10 Toronto time = exactly midpoint of the trip
	loc, _ := time.LoadLocation("America/Toronto")
	midpoint := time.Date(2026, 3, 3, 9, 10, 0, 0, loc)

	positions := gtfsstore.SimulatePositions(midpoint, store)

	if len(positions) != 1 {
		t.Fatalf("expected 1 position, got %d", len(positions))
	}
	p := positions[0]
	if p.VehicleID != "T001" {
		t.Errorf("expected VehicleID T001, got %s", p.VehicleID)
	}

	wantLat := 43.6453 + 0.5*(43.6200-43.6453)
	wantLon := -79.3806 + 0.5*(-79.4900-(-79.3806))
	if math.Abs(p.Lat-wantLat) > 0.0001 {
		t.Errorf("lat: want %.4f, got %.4f", wantLat, p.Lat)
	}
	if math.Abs(p.Lon-wantLon) > 0.0001 {
		t.Errorf("lon: want %.4f, got %.4f", wantLon, p.Lon)
	}
}

func TestSimulatePositions_BeforeTrip(t *testing.T) {
	store, err := gtfsstore.NewStaticStore(buildSimZip(t))
	if err != nil {
		t.Fatal(err)
	}
	loc, _ := time.LoadLocation("America/Toronto")
	before := time.Date(2026, 3, 3, 8, 0, 0, 0, loc)
	positions := gtfsstore.SimulatePositions(before, store)
	if len(positions) != 0 {
		t.Errorf("expected 0 positions before trip start, got %d", len(positions))
	}
}

func TestSimulatePositions_AfterTrip(t *testing.T) {
	store, err := gtfsstore.NewStaticStore(buildSimZip(t))
	if err != nil {
		t.Fatal(err)
	}
	loc, _ := time.LoadLocation("America/Toronto")
	after := time.Date(2026, 3, 3, 10, 0, 0, 0, loc)
	positions := gtfsstore.SimulatePositions(after, store)
	if len(positions) != 0 {
		t.Errorf("expected 0 positions after trip end, got %d", len(positions))
	}
}
