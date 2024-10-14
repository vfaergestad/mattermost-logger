package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Duration is a custom type that wraps time.Duration to support JSON unmarshalling from strings.
type Duration time.Duration

// UnmarshalJSON parses a JSON string into a Duration.
func (d *Duration) UnmarshalJSON(b []byte) error {
	// Remove quotes from the JSON string
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return fmt.Errorf("Duration should be a string, got %s", string(b))
	}

	// Parse the duration string
	duration, err := time.ParseDuration(s)
	if err != nil {
		return fmt.Errorf("invalid duration '%s': %v", s, err)
	}

	*d = Duration(duration)
	return nil
}

type Config struct {
	MattermostDomain      string   `json:"mattermost_domain"`
	MattermostPort        int      `json:"mattermost_port,omitempty"` // Optional: defaults to 443 for tls, else 80
	UseTLS                bool     `json:"use_tls"`
	AuthToken             string   `json:"auth_token"`
	ChannelIDs            []string `json:"channel_ids"`
	OutputDir             string   `json:"output_dir,omitempty"`               // New: Directory for output files
	OutputFilePrefix      string   `json:"output_file_prefix,omitempty"`       // New: Prefix for output files
	InsecureSkipTLSVerify bool     `json:"insecure_skip_tls_verify,omitempty"` // Optional: defaults to false
	LogLevel              string   `json:"log_level,omitempty"`                // Optional: DEBUG, INFO, WARN, ERROR
	LogFile               string   `json:"log_file,omitempty"`                 // Optional: defaults to logs/app.log
	RotationInterval      Duration `json:"rotation_interval,omitempty"`
}

func LoadConfig(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Error opening config file: %v\n", err)
		return nil, err
	}
	defer file.Close()

	cfg := &Config{}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(cfg)
	if err != nil {
		fmt.Printf("Error decoding config file: %v\n", err)
		return nil, err
	}

	// Set default MattermostPort if not specified
	if cfg.MattermostPort == 0 {
		if cfg.UseTLS {
			cfg.MattermostPort = 443 // Default HTTPS port
		} else {
			cfg.MattermostPort = 80 // Default HTTP port
		}
	}

	// Set default OutputDir if not specified
	if cfg.OutputDir == "" {
		cfg.OutputDir = "out" // Default output directory
	}

	// Set default OutputFilePrefix if not specified
	if cfg.OutputFilePrefix == "" {
		cfg.OutputFilePrefix = "messages" // Default file prefix
	}

	// Set default RotationInterval if not specified
	if cfg.RotationInterval == 0 {
		cfg.RotationInterval = Duration(24 * time.Hour) // Default rotation interval
	}

	// Ensure OutputDir exists
	path := filepath.Join(".", cfg.OutputDir)
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		fmt.Printf("Failed to create output directory %s: %v\n", cfg.OutputDir, err)
		return nil, err
	}

	// Set default LogLevel if not specified
	if cfg.LogLevel == "" {
		cfg.LogLevel = "INFO"
	}

	// Set default LogFile if not specified
	if cfg.LogFile == "" {
		cfg.LogFile = "logs/app.log"
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		fmt.Printf("Configuration validation error: %v\n", err)
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.MattermostDomain == "" {
		return fmt.Errorf("mattermost_domain is required")
	}
	if c.AuthToken == "" {
		return fmt.Errorf("auth_token is required")
	}
	if len(c.ChannelIDs) == 0 {
		return fmt.Errorf("at least one channel_id is required")
	}
	return nil
}
