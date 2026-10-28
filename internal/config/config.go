// Package config loads and validates application configuration.
package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Port int `yaml:"port"`
	} `yaml:"server"`
	TFL struct {
		BaseURL           string   `yaml:"base_url"`
		MonitoredLines    []string `yaml:"monitored_lines"`
		MonitoredStations []string `yaml:"monitored_stations"`
		AppKey            string   `yaml:"-"`
	} `yaml:"tfl"`
}

// Load reads the YAML file at path and applies environment overrides
// (PORT, TFL_BASE_URL, TFL_APP_KEY).
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
	c.TFL.AppKey = os.Getenv("TFL_APP_KEY")
	if c.Server.Port == 0 {
		c.Server.Port = 8080
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
	if c.TFL.BaseURL == "" {
		return errors.New("tfl.base_url is required")
	}
	return nil
}
