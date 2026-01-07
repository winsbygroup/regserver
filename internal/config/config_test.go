package config_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"winsbygroup.com/regserver/internal/config"
)

func TestLoad(t *testing.T) {
	// Helper to clear env vars before each test
	clearEnvVars := func() {
		os.Unsetenv("DB_PATH")
		os.Unsetenv("API_KEY")
		os.Unsetenv("REGISTRATION_SECRET")
	}

	t.Run("returns defaults when config file does not exist", func(t *testing.T) {
		clearEnvVars()

		cfg, err := config.Load("nonexistent.yaml")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if cfg.DBPath != "./registrations.db" {
			t.Errorf("expected DBPath './registrations.db', got %q", cfg.DBPath)
		}
		if cfg.ReadTimeout != 5*time.Second {
			t.Errorf("expected ReadTimeout 5s, got %v", cfg.ReadTimeout)
		}
		if cfg.WriteTimeout != 10*time.Second {
			t.Errorf("expected WriteTimeout 10s, got %v", cfg.WriteTimeout)
		}
		if cfg.IdleTimeout != 120*time.Second {
			t.Errorf("expected IdleTimeout 120s, got %v", cfg.IdleTimeout)
		}
	})

	t.Run("loads values from YAML file", func(t *testing.T) {
		clearEnvVars()

		// Create temp config file
		tmpDir := t.TempDir()
		cfgPath := filepath.Join(tmpDir, "config.yaml")
		yamlContent := `
addr: ":9090"
db_path: "/data/test.db"
api_key: "yaml-api-key"
registration_secret: "yaml-secret"
read_timeout: 15s
write_timeout: 30s
idle_timeout: 60s
`
		if err := os.WriteFile(cfgPath, []byte(yamlContent), 0644); err != nil {
			t.Fatalf("failed to write config file: %v", err)
		}

		cfg, err := config.Load(cfgPath)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if cfg.Addr != ":9090" {
			t.Errorf("expected Addr ':9090', got %q", cfg.Addr)
		}
		if cfg.DBPath != "/data/test.db" {
			t.Errorf("expected DBPath '/data/test.db', got %q", cfg.DBPath)
		}
		if cfg.APIKey != "yaml-api-key" {
			t.Errorf("expected APIKey 'yaml-api-key', got %q", cfg.APIKey)
		}
		if cfg.RegistrationSecret != "yaml-secret" {
			t.Errorf("expected RegistrationSecret 'yaml-secret', got %q", cfg.RegistrationSecret)
		}
		if cfg.ReadTimeout != 15*time.Second {
			t.Errorf("expected ReadTimeout 15s, got %v", cfg.ReadTimeout)
		}
		if cfg.WriteTimeout != 30*time.Second {
			t.Errorf("expected WriteTimeout 30s, got %v", cfg.WriteTimeout)
		}
		if cfg.IdleTimeout != 60*time.Second {
			t.Errorf("expected IdleTimeout 60s, got %v", cfg.IdleTimeout)
		}
	})

	t.Run("env vars override defaults when no config file", func(t *testing.T) {
		clearEnvVars()
		os.Setenv("DB_PATH", "/env/path.db")
		os.Setenv("API_KEY", "env-api-key")
		os.Setenv("REGISTRATION_SECRET", "env-secret")
		defer clearEnvVars()

		cfg, err := config.Load("nonexistent.yaml")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if cfg.DBPath != "/env/path.db" {
			t.Errorf("expected DBPath '/env/path.db', got %q", cfg.DBPath)
		}
		if cfg.APIKey != "env-api-key" {
			t.Errorf("expected APIKey 'env-api-key', got %q", cfg.APIKey)
		}
		if cfg.RegistrationSecret != "env-secret" {
			t.Errorf("expected RegistrationSecret 'env-secret', got %q", cfg.RegistrationSecret)
		}
	})

	t.Run("env vars override YAML values", func(t *testing.T) {
		clearEnvVars()

		// Create temp config file
		tmpDir := t.TempDir()
		cfgPath := filepath.Join(tmpDir, "config.yaml")
		yamlContent := `
db_path: "/yaml/path.db"
api_key: "yaml-api-key"
registration_secret: "yaml-secret"
`
		if err := os.WriteFile(cfgPath, []byte(yamlContent), 0644); err != nil {
			t.Fatalf("failed to write config file: %v", err)
		}

		// Set env vars that should override
		os.Setenv("DB_PATH", "/env/override.db")
		os.Setenv("API_KEY", "env-override-key")
		os.Setenv("REGISTRATION_SECRET", "env-override-secret")
		defer clearEnvVars()

		cfg, err := config.Load(cfgPath)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if cfg.DBPath != "/env/override.db" {
			t.Errorf("expected DBPath '/env/override.db', got %q", cfg.DBPath)
		}
		if cfg.APIKey != "env-override-key" {
			t.Errorf("expected APIKey 'env-override-key', got %q", cfg.APIKey)
		}
		if cfg.RegistrationSecret != "env-override-secret" {
			t.Errorf("expected RegistrationSecret 'env-override-secret', got %q", cfg.RegistrationSecret)
		}
	})

	t.Run("partial env var override", func(t *testing.T) {
		clearEnvVars()

		// Create temp config file
		tmpDir := t.TempDir()
		cfgPath := filepath.Join(tmpDir, "config.yaml")
		yamlContent := `
db_path: "/yaml/path.db"
api_key: "yaml-api-key"
registration_secret: "yaml-secret"
`
		if err := os.WriteFile(cfgPath, []byte(yamlContent), 0644); err != nil {
			t.Fatalf("failed to write config file: %v", err)
		}

		// Only override DB_PATH
		os.Setenv("DB_PATH", "/env/only-db.db")
		defer clearEnvVars()

		cfg, err := config.Load(cfgPath)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// DB_PATH should be overridden
		if cfg.DBPath != "/env/only-db.db" {
			t.Errorf("expected DBPath '/env/only-db.db', got %q", cfg.DBPath)
		}
		// Others should come from YAML
		if cfg.APIKey != "yaml-api-key" {
			t.Errorf("expected APIKey 'yaml-api-key', got %q", cfg.APIKey)
		}
		if cfg.RegistrationSecret != "yaml-secret" {
			t.Errorf("expected RegistrationSecret 'yaml-secret', got %q", cfg.RegistrationSecret)
		}
	})

	t.Run("returns error for invalid YAML", func(t *testing.T) {
		clearEnvVars()

		// Create temp config file with invalid YAML
		tmpDir := t.TempDir()
		cfgPath := filepath.Join(tmpDir, "config.yaml")
		invalidYAML := `
addr: ":9090"
  invalid indentation
db_path: "/data/test.db"
`
		if err := os.WriteFile(cfgPath, []byte(invalidYAML), 0644); err != nil {
			t.Fatalf("failed to write config file: %v", err)
		}

		_, err := config.Load(cfgPath)
		if err == nil {
			t.Error("expected error for invalid YAML, got nil")
		}
	})
}
