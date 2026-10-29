package config

import "testing"

func TestParse(t *testing.T) {
	t.Setenv("TFL_APP_KEY", "k")
	t.Setenv("PORT", "9000")
	t.Setenv("DASHBOARD_REFRESH_INTERVAL_SECONDS", "45")
	c, err := Parse([]byte("tfl:\n  base_url: http://x\n  monitored_lines: [victoria]\n"))
	if err != nil {
		t.Fatal(err)
	}
	if c.Server.Port != 9000 || c.TFL.AppKey != "k" || c.Dashboard.RefreshIntervalSeconds != 45 || len(c.TFL.MonitoredLines) != 1 {
		t.Fatalf("unexpected: %+v", c)
	}
}

func TestParseDefaultRefreshInterval(t *testing.T) {
	t.Setenv("DASHBOARD_REFRESH_INTERVAL_SECONDS", "")
	c, err := Parse([]byte("tfl:\n  base_url: http://x\n"))
	if err != nil {
		t.Fatal(err)
	}
	if c.Dashboard.RefreshIntervalSeconds != 30 {
		t.Fatalf("refresh interval: got %d, want 30", c.Dashboard.RefreshIntervalSeconds)
	}
}

func TestParseConfiguredRefreshInterval(t *testing.T) {
	t.Setenv("DASHBOARD_REFRESH_INTERVAL_SECONDS", "")
	c, err := Parse([]byte("dashboard:\n  refresh_interval_seconds: 45\ntfl:\n  base_url: http://x\n"))
	if err != nil {
		t.Fatal(err)
	}
	if c.Dashboard.RefreshIntervalSeconds != 45 {
		t.Fatalf("refresh interval: got %d, want 45", c.Dashboard.RefreshIntervalSeconds)
	}
}

func TestParseMissingBaseURL(t *testing.T) {
	if _, err := Parse([]byte("server:\n  port: 80\n")); err == nil {
		t.Fatal("expected error")
	}
}

func TestParseInvalidRefreshInterval(t *testing.T) {
	t.Setenv("DASHBOARD_REFRESH_INTERVAL_SECONDS", "")
	if _, err := Parse([]byte("dashboard:\n  refresh_interval_seconds: -1\ntfl:\n  base_url: http://x\n")); err == nil {
		t.Fatal("expected error")
	}
}
