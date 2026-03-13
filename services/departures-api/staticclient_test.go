package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/teclara/railsix/shared/models"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func TestStaticClientInvalidatesCachesWhenGTFSVersionChanges(t *testing.T) {
	var (
		mu         sync.Mutex
		version    = "v1"
		routeShort = "LW"
		routeCalls int
	)

	client := NewStaticClient("https://gtfs-static.test")
	client.client = &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			mu.Lock()
			currentVersion := version
			currentRouteShort := routeShort
			if req.URL.Path == "/routes/route-1" {
				routeCalls++
			}
			mu.Unlock()

			headers := make(http.Header)
			headers.Set("Content-Type", "application/json")
			headers.Set("X-GTFS-Version", currentVersion)

			var body []byte
			switch req.URL.Path {
			case "/routes/route-1":
				data, err := json.Marshal(models.Route{
					ID:        "route-1",
					ShortName: currentRouteShort,
					LongName:  "Lakeshore West",
				})
				if err != nil {
					t.Fatalf("marshal route: %v", err)
				}
				body = data
			case "/stops":
				body = []byte("[]")
			default:
				return &http.Response{
					StatusCode: http.StatusNotFound,
					Header:     headers,
					Body:       io.NopCloser(strings.NewReader(`{"error":"not found"}`)),
				}, nil
			}

			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     headers,
				Body:       io.NopCloser(bytes.NewReader(body)),
			}, nil
		}),
	}

	route, ok := client.GetRoute("route-1")
	if !ok {
		t.Fatal("expected initial route lookup to succeed")
	}
	if route.ShortName != "LW" {
		t.Fatalf("expected cached route short name LW, got %q", route.ShortName)
	}

	mu.Lock()
	version = "v2"
	routeShort = "KI"
	mu.Unlock()

	if _, err := client.GetStops(); err != nil {
		t.Fatalf("expected version-refreshing stops request to succeed: %v", err)
	}

	route, ok = client.GetRoute("route-1")
	if !ok {
		t.Fatal("expected route lookup after version change to succeed")
	}
	if route.ShortName != "KI" {
		t.Fatalf("expected cache invalidation to refresh route short name to KI, got %q (version=%q)", route.ShortName, client.gtfsVersion)
	}

	mu.Lock()
	defer mu.Unlock()
	if routeCalls != 2 {
		t.Fatalf("expected route endpoint to be called twice, got %d", routeCalls)
	}
}
