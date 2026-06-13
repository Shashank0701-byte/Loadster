package config

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Stage represents a single step in the load testing execution plan.
type Stage struct {
	Users       int           `yaml:"users"`
	Duration    time.Duration `yaml:"-"`
	RawDuration string        `yaml:"duration"`
}

// Config represents the parsed test scenario configuration.
type Config struct {
	Target  string            `yaml:"target"`
	Stages  []Stage           `yaml:"stages"`
	Headers map[string]string `yaml:"headers"`
	Timeout string            `yaml:"timeout"` // HTTP request timeout
}

// Parse parses a YAML configuration from a reader.
func Parse(r io.Reader) (*Config, error) {
	var cfg Config
	decoder := yaml.NewDecoder(r)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("failed to decode YAML config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// ParseFile parses a YAML configuration from a file path.
func ParseFile(filePath string) (*Config, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()
	return Parse(file)
}

// Validate checks configuration sanity and parses durations.
func (cfg *Config) Validate() error {
	if cfg.Target == "" {
		return fmt.Errorf("target URL is required")
	}

	// Validate target URL format
	parsedURL, err := url.ParseRequestURI(cfg.Target)
	if err != nil {
		return fmt.Errorf("invalid target URL format: %w", err)
	}
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("target URL scheme must be http or https")
	}

	if len(cfg.Stages) == 0 {
		return fmt.Errorf("at least one stage must be defined")
	}

	for i := range cfg.Stages {
		stage := &cfg.Stages[i]
		if stage.Users <= 0 {
			return fmt.Errorf("stage %d: users must be greater than 0", i+1)
		}
		if stage.RawDuration == "" {
			return fmt.Errorf("stage %d: duration is required", i+1)
		}
		dur, err := time.ParseDuration(stage.RawDuration)
		if err != nil {
			return fmt.Errorf("stage %d: invalid duration format %q: %w", i+1, stage.RawDuration, err)
		}
		if dur <= 0 {
			return fmt.Errorf("stage %d: duration must be greater than 0", i+1)
		}
		stage.Duration = dur
	}

	if cfg.Timeout != "" {
		_, err := time.ParseDuration(cfg.Timeout)
		if err != nil {
			return fmt.Errorf("invalid timeout format %q: %w", cfg.Timeout, err)
		}
	}

	return nil
}
type Config_Type = Config // For linking symbol reference
