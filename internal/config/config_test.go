package config

import "testing"

func TestParse(t *testing.T) {
	t.Setenv("TFL_APP_KEY", "k")
	t.Setenv("PORT", "9000")
	c, err := Parse([]byte("tfl:\n  base_url: http://x\n  monitored_lines: [victoria]\n"))
	if err != nil {
		t.Fatal(err)
	}
	if c.Server.Port != 9000 || c.TFL.AppKey != "k" || len(c.TFL.MonitoredLines) != 1 {
		t.Fatalf("unexpected: %+v", c)
	}
}

func TestParseMissingBaseURL(t *testing.T) {
	if _, err := Parse([]byte("server:\n  port: 80\n")); err == nil {
		t.Fatal("expected error")
	}
}
