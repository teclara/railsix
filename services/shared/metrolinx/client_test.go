package metrolinx

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestNewClient(t *testing.T) {
	c := NewClient("https://example.com", "test-key")
	if c == nil {
		t.Fatal("NewClient returned nil")
	}
}

func TestParseMetrolinxTime_Valid(t *testing.T) {
	got := parseMetrolinxTime("2026-03-10 14:35:00")
	want := "14:35"
	if got != want {
		t.Errorf("parseMetrolinxTime(%q) = %q, want %q", "2026-03-10 14:35:00", got, want)
	}
}

func TestParseMetrolinxTime_Invalid(t *testing.T) {
	got := parseMetrolinxTime("not-a-time")
	want := "--:--"
	if got != want {
		t.Errorf("parseMetrolinxTime(%q) = %q, want %q", "not-a-time", got, want)
	}
}

func TestParseMetrolinxTime_Empty(t *testing.T) {
	got := parseMetrolinxTime("")
	want := "--:--"
	if got != want {
		t.Errorf("parseMetrolinxTime(%q) = %q, want %q", "", got, want)
	}
}

func TestFetchRejectsOversizedResponses(t *testing.T) {
	oversizedBody := bytes.Repeat([]byte("a"), 10*1024*1024+1)
	client := NewClient("https://metrolinx.test", "test-key")
	client.httpClient = &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader(oversizedBody)),
			}, nil
		}),
	}
	data, err := client.Fetch(context.Background(), "/Gtfs/Feed/Alerts")
	if err == nil {
		t.Fatal("expected oversized response to fail")
	}
	if data != nil {
		t.Fatalf("expected no data on oversized response, got %d bytes", len(data))
	}
	if !strings.Contains(err.Error(), "exceeds") {
		t.Fatalf("expected oversize error, got %v", err)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}
