package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// CLIBackend represents the type of CLI to use
type CLIBackend string

const (
	CLIBackendClaude CLIBackend = "claude"
	CLIBackendCodex  CLIBackend = "codex"
)

// CLIConfig holds CLI-specific configuration
type CLIConfig struct {
	Backend   CLIBackend `json:"cli"`
	Command   string     `json:"command,omitempty"`   // Override command path
	Model     string     `json:"model,omitempty"`     // Model to use
	ExtraArgs []string   `json:"extra_args,omitempty"` // Additional CLI arguments
}

// DefaultCLIConfig returns the default CLI configuration (Claude)
func DefaultCLIConfig() CLIConfig {
	return CLIConfig{
		Backend: CLIBackendClaude,
	}
}

// ProjectConfig represents the .ralph-config.json file
type ProjectConfig struct {
	CLI   CLIBackend `json:"cli,omitempty"`
	Model string     `json:"model,omitempty"`
}

// LoadCLIConfig loads CLI configuration with the following precedence:
// 1. Explicit overrides (from flags)
// 2. Environment variables (RALPH_CLI, RALPH_MODEL)
// 3. .ralph-config.json in project root
// 4. Defaults
func LoadCLIConfig(flagCLI, flagModel string) CLIConfig {
	config := DefaultCLIConfig()

	// Load from project config file (lowest priority)
	if projectConfig, err := loadProjectConfig(); err == nil {
		if projectConfig.CLI != "" {
			config.Backend = projectConfig.CLI
		}
		if projectConfig.Model != "" {
			config.Model = projectConfig.Model
		}
	}

	// Environment variables (medium priority)
	if envCLI := os.Getenv("RALPH_CLI"); envCLI != "" {
		config.Backend = CLIBackend(envCLI)
	}
	if envModel := os.Getenv("RALPH_MODEL"); envModel != "" {
		config.Model = envModel
	}

	// Command-line flags (highest priority)
	if flagCLI != "" {
		config.Backend = CLIBackend(flagCLI)
	}
	if flagModel != "" {
		config.Model = flagModel
	}

	return config
}

// loadProjectConfig loads .ralph-config.json from the current directory
func loadProjectConfig() (*ProjectConfig, error) {
	configPath := filepath.Join(".", ".ralph-config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config ProjectConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// IsValid checks if the CLI backend is valid
func (b CLIBackend) IsValid() bool {
	return b == CLIBackendClaude || b == CLIBackendCodex
}

// String returns the string representation of the CLI backend
func (b CLIBackend) String() string {
	return string(b)
}
