package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	MattermostDomain      string   `json:"mattermost_domain"`
	MattermostPort        int      `json:"mattermost_port,omitempty"` // Optional: defaults to 443 for tls, else 80
	UseTLS                bool     `json:"use_tls"`
	AuthToken             string   `json:"auth_token"`
	ChannelIDs            []string `json:"channel_ids"`
	OutputFile            string   `json:"output_file,omitempty"`              // Optional: defaults to messages.json
	InsecureSkipTLSVerify bool     `json:"insecure_skip_tls_verify,omitempty"` // Optional: defaults to false
	LogLevel              string   `json:"log_level,omitempty"`                // Optional: DEBUG, INFO, WARN, ERROR
	LogFile               string   `json:"log_file,omitempty"`                 // Optional: defaults to logs/app.log
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

	// Set default OutputFile if not specified
	if cfg.OutputFile == "" {
		cfg.OutputFile = "messages.json"
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
