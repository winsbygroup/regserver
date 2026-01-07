package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds all configuration values
type Config struct {
	Addr               string        `yaml:"addr"`
	DBPath             string        `yaml:"db_path"`
	APIKey             string        `yaml:"api_key"`
	RegistrationSecret string        `yaml:"registration_secret"`
	ReadTimeout        time.Duration `yaml:"read_timeout"`
	WriteTimeout       time.Duration `yaml:"write_timeout"`
	IdleTimeout        time.Duration `yaml:"idle_timeout"`

	DBPathSource string // where DBPath was set from: "default", "yaml file", or "env var"
	DemoMode     bool   // load sample data on new database (set via -demo flag)
}

// Load loads configuration from YAML file and overrides with env vars if present
func Load(path string) (*Config, error) {
	// Defaults
	cfg := &Config{
		Addr:         ":8080",
		DBPath:       "./registrations.db",
		DBPathSource: "default",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Load from YAML if file exists
	if f, err := os.Open(path); err == nil {
		defer f.Close()
		prevDBPath := cfg.DBPath
		decoder := yaml.NewDecoder(f)
		if err := decoder.Decode(cfg); err != nil {
			return nil, err
		}
		if cfg.DBPath != prevDBPath {
			cfg.DBPathSource = "yaml file"
		}
	}

	// Override with environment variables
	if v := os.Getenv("PORT"); v != "" {
		cfg.Addr = ":" + v
	}
	if v := os.Getenv("DB_PATH"); v != "" {
		cfg.DBPath = v
		cfg.DBPathSource = "env var"
	}
	if v := os.Getenv("API_KEY"); v != "" {
		cfg.APIKey = v
	}
	if v := os.Getenv("REGISTRATION_SECRET"); v != "" {
		cfg.RegistrationSecret = v
	}

	return cfg, nil
}
