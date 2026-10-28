package tfl

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLineStatusesAppKey(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/line/victoria,northern/status" || r.URL.Query().Get("app_key") != "k" {
			t.Errorf("bad request %s", r.URL)
		}
		w.Write([]byte(`[{"id":"victoria","name":"Victoria","lineStatuses":[{"statusSeverityDescription":"Good Service"}]}]`))
	}))
	defer srv.Close()
	got, err := NewClient(srv.URL, "k").LineStatuses(context.Background(), []string{"victoria", "northern"})
	if err != nil || len(got) != 1 || got[0].LineStatuses[0].StatusSeverityDescription != "Good Service" {
		t.Fatalf("got %+v err %v", got, err)
	}
}

func TestArrivalsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	defer srv.Close()
	if _, err := NewClient(srv.URL, "").Arrivals(context.Background(), "x"); err == nil {
		t.Fatal("expected error")
	}
}
