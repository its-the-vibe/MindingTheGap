package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/its-the-vibe/MindingTheGap/internal/config"
	"github.com/its-the-vibe/MindingTheGap/internal/tfl"
)

type fake struct {
	arrivalCalls []string
}

func (*fake) LineStatuses(context.Context, []string) ([]tfl.LineStatus, error) {
	return []tfl.LineStatus{{ID: "victoria", Name: "Victoria"}}, nil
}
func (f *fake) Arrivals(_ context.Context, line, stationID string) ([]tfl.Arrival, error) {
	f.arrivalCalls = append(f.arrivalCalls, line+"/"+stationID)
	return []tfl.Arrival{{LineName: line, TimeToStation: 60}}, nil
}

func TestHandlers(t *testing.T) {
	cfg := &config.Config{}
	cfg.Dashboard.RefreshIntervalSeconds = 30
	cfg.TFL.MonitoredStations = []string{"victoria/X"}
	h := New(cfg, &fake{}).Handler()

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/api/v1/status", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"Victoria"`) {
		t.Fatalf("status: %d %s", rec.Code, rec.Body)
	}
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "Tube Dashboard") {
		t.Fatalf("index: %d", rec.Code)
	}
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/nope", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestIndexUsesConfiguredRefreshInterval(t *testing.T) {
	cfg := &config.Config{}
	cfg.Dashboard.RefreshIntervalSeconds = 45
	rec := httptest.NewRecorder()
	New(cfg, &fake{}).Handler().ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))

	if !strings.Contains(rec.Body.String(), "setInterval(load,45000)") {
		t.Fatalf("index does not contain configured refresh interval: %s", rec.Body)
	}
}

func TestStatusQueriesEachMonitoredLineAndStation(t *testing.T) {
	cfg := &config.Config{}
	cfg.TFL.MonitoredStations = []string{"northern/940GZZLUOXC", "victoria/940GZZLUOXC"}
	src := &fake{}
	rec := httptest.NewRecorder()
	New(cfg, src).Handler().ServeHTTP(rec, httptest.NewRequest("GET", "/api/v1/status", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status: %d %s", rec.Code, rec.Body)
	}
	if strings.Join(src.arrivalCalls, ",") != "northern/940GZZLUOXC,victoria/940GZZLUOXC" {
		t.Fatalf("arrival calls: %v", src.arrivalCalls)
	}
	var got StatusResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Arrivals) != 2 || got.Arrivals["northern/940GZZLUOXC"][0].LineName != "northern" ||
		got.Arrivals["victoria/940GZZLUOXC"][0].LineName != "victoria" {
		t.Fatalf("arrivals: %+v", got.Arrivals)
	}
}
