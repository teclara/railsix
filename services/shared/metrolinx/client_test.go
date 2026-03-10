package metrolinx

import (
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
