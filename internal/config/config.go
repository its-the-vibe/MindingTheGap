// Package config loads and validates application configuration.
package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

const maxRefreshIntervalSeconds = 2147483

type Config struct {
	Server struct {
		Port int `yaml:"port"`
	} `yaml:"server"`
	Dashboard struct {
		RefreshIntervalSeconds int `yaml:"refresh_interval_seconds"`
	} `yaml:"dashboard"`
	TFL struct {
		BaseURL           string   `yaml:"base_url"`
		MonitoredLines    []string `yaml:"monitored_lines"`
		MonitoredStations []string `yaml:"monitored_stations"`
		AppKey            string   `yaml:"-"`
	} `yaml:"tfl"`
}

// Load reads the YAML file at path and applies environment overrides
// (PORT, TFL_BASE_URL, TFL_APP_KEY, DASHBOARD_REFRESH_INTERVAL_SECONDS).
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	return Parse(data)
}

func Parse(data []byte) (*Config, error) {
	var c Config
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	if v := os.Getenv("PORT"); v != "" {
		p, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("invalid PORT %q: %w", v, err)
		}
		c.Server.Port = p
	}
	if v := os.Getenv("TFL_BASE_URL"); v != "" {
		c.TFL.BaseURL = v
	}
	if v := os.Getenv("DASHBOARD_REFRESH_INTERVAL_SECONDS"); v != "" {
		interval, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("invalid DASHBOARD_REFRESH_INTERVAL_SECONDS %q: %w", v, err)
		}
		if interval < 1 {
			return nil, errors.New("DASHBOARD_REFRESH_INTERVAL_SECONDS must be positive")
		}
		c.Dashboard.RefreshIntervalSeconds = interval
	}
	c.TFL.AppKey = os.Getenv("TFL_APP_KEY")
	if c.Server.Port == 0 {
		c.Server.Port = 8080
	}
	if c.Dashboard.RefreshIntervalSeconds == 0 {
		c.Dashboard.RefreshIntervalSeconds = 30
	}
	if err := c.validate(); err != nil {
		return nil, err
	}
	return &c, nil
}

func (c *Config) validate() error {
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port %d", c.Server.Port)
	}
	if c.Dashboard.RefreshIntervalSeconds < 1 || c.Dashboard.RefreshIntervalSeconds > maxRefreshIntervalSeconds {
		return fmt.Errorf("dashboard.refresh_interval_seconds must be between 1 and %d", maxRefreshIntervalSeconds)
	}
	if c.TFL.BaseURL == "" {
		return errors.New("tfl.base_url is required")
	}
	return nil
}
