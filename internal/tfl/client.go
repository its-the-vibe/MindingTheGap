// Package tfl is a minimal client for the TfL Unified API.
package tfl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	baseURL string
	appKey  string
	http    *http.Client
}

func NewClient(baseURL, appKey string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		appKey:  appKey,
		http:    &http.Client{Timeout: 10 * time.Second},
	}
}

type LineStatus struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	LineStatuses []struct {
		StatusSeverityDescription string `json:"statusSeverityDescription"`
		Reason                    string `json:"reason,omitempty"`
	} `json:"lineStatuses"`
}

type Arrival struct {
	ID              string    `json:"id"`
	StationName     string    `json:"stationName"`
	LineName        string    `json:"lineName"`
	PlatformName    string    `json:"platformName"`
	DestinationName string    `json:"destinationName"`
	TowardsText     string    `json:"towards"`
	Direction       string    `json:"direction"`
	TimeToStation   int       `json:"timeToStation"`
	ExpectedArrival time.Time `json:"expectedArrival"`
}

func (c *Client) get(ctx context.Context, path string, out any) error {
	u := c.baseURL + path
	if c.appKey != "" {
		u += "?" + url.Values{"app_key": {c.appKey}}.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "curl/8.7.1")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("tfl api %s: status %d", path, resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

// LineStatuses calls GET /line/{ids}/status.
func (c *Client) LineStatuses(ctx context.Context, ids []string) ([]LineStatus, error) {
	var out []LineStatus
	if len(ids) == 0 {
		return out, nil
	}
	escaped := make([]string, len(ids))
	for i, id := range ids {
		escaped[i] = url.PathEscape(id)
	}
	err := c.get(ctx, "/line/"+strings.Join(escaped, ",")+"/status", &out)
	return out, err
}

// Arrivals calls GET /line/{line}/arrivals/{stationID}.
func (c *Client) Arrivals(ctx context.Context, line, stationID string) ([]Arrival, error) {
	var out []Arrival
	err := c.get(ctx, "/line/"+url.PathEscape(line)+"/arrivals/"+url.PathEscape(stationID), &out)
	return out, err
}
