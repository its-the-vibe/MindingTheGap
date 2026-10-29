package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/its-the-vibe/MindingTheGap/internal/config"
	"github.com/its-the-vibe/MindingTheGap/internal/tfl"
)

type fake struct{}

func (fake) LineStatuses(context.Context, []string) ([]tfl.LineStatus, error) {
	return []tfl.LineStatus{{ID: "victoria", Name: "Victoria"}}, nil
}
func (fake) Arrivals(context.Context, string) ([]tfl.Arrival, error) {
	return []tfl.Arrival{{LineName: "Victoria", TimeToStation: 60}}, nil
}

func TestHandlers(t *testing.T) {
	cfg := &config.Config{}
	cfg.TFL.MonitoredStations = []string{"X"}
	h := New(cfg, fake{}).Handler()

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
