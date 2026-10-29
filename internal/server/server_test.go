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
	arrivals     []tfl.Arrival
}

func (*fake) LineStatuses(context.Context, []string) ([]tfl.LineStatus, error) {
	return []tfl.LineStatus{{ID: "victoria", Name: "Victoria"}}, nil
}
func (f *fake) Arrivals(_ context.Context, line, stationID string) ([]tfl.Arrival, error) {
	f.arrivalCalls = append(f.arrivalCalls, line+"/"+stationID)
	if f.arrivals != nil {
		return append([]tfl.Arrival(nil), f.arrivals...), nil
	}
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
	if len(got.ArrivalsByDirection["northern/940GZZLUOXC"]) != 1 ||
		len(got.ArrivalsByDirection["northern/940GZZLUOXC"]["Unknown direction"]) != 1 {
		t.Fatalf("single-direction arrivals: %+v", got.ArrivalsByDirection["northern/940GZZLUOXC"])
	}
}

func TestStatusGroupsArrivalsByDirectionAndKeepsFlatArrivals(t *testing.T) {
	cfg := &config.Config{}
	cfg.TFL.MonitoredStations = []string{"northern/940GZZLUOXC"}
	src := &fake{arrivals: []tfl.Arrival{
		{LineName: "northern", Direction: "southbound", TowardsText: "Morden", TimeToStation: 420},
		{LineName: "northern", Direction: "southbound", TowardsText: "Battersea Power", TimeToStation: 180},
		{LineName: "northern", Direction: "northbound", TowardsText: "Edgware", TimeToStation: 60},
		{LineName: "northern", Direction: " northbound ", TowardsText: "High Barnet", TimeToStation: 120},
		{LineName: "northern", TowardsText: "Morden", TimeToStation: 30},
	}}
	rec := httptest.NewRecorder()
	New(cfg, src).Handler().ServeHTTP(rec, httptest.NewRequest("GET", "/api/v1/status", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status: %d %s", rec.Code, rec.Body)
	}
	var got StatusResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	station := "northern/940GZZLUOXC"
	if got.Arrivals[station][0].TimeToStation != 30 || len(got.Arrivals[station]) != 5 {
		t.Fatalf("flat arrivals: %+v", got.Arrivals[station])
	}
	groups := got.ArrivalsByDirection[station]
	if len(groups) != 3 {
		t.Fatalf("direction groups: %+v", groups)
	}
	if got := groups["southbound"]; len(got) != 2 || got[0].TimeToStation != 180 || got[1].TimeToStation != 420 {
		t.Fatalf("southbound arrivals: %+v", got)
	}
	if got := groups["northbound"]; len(got) != 2 || got[0].TimeToStation != 60 || got[1].TimeToStation != 120 {
		t.Fatalf("northbound arrivals: %+v", got)
	}
	if got := groups["Unknown direction"]; len(got) != 1 || got[0].TimeToStation != 30 {
		t.Fatalf("unknown direction arrivals: %+v", got)
	}
}

func TestIndexRendersDirectionGroups(t *testing.T) {
	cfg := &config.Config{}
	rec := httptest.NewRecorder()
	New(cfg, &fake{}).Handler().ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))

	if !strings.Contains(rec.Body.String(), "d.arrivalsByDirection") ||
		!strings.Contains(rec.Body.String(), "el('h3',null,direction)") ||
		!strings.Contains(rec.Body.String(), "a.towards||''") ||
		strings.Contains(rec.Body.String(), "a.destinationName") {
		t.Fatalf("index does not render direction groups: %s", rec.Body)
	}
}
