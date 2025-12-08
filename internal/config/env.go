// Package config provides configuration management for Sleepship.
// It handles loading and merging configurations from multiple sources including
// environment variables, CLI flags, and default values.
package config

import (
	"os"
	"strconv"
	"strings"
)

// EnvConfig represents configuration loaded from environment variables
type EnvConfig struct {
	ProjectDir      string
	DefaultTaskFile string
	MaxRetries      int
	LogDir          string
	StartFrom       int
	ClaudeFlags     []string
}

// LoadFromEnv loads configuration from environment variables
// Environment variables should be prefixed with SLEEPSHIP_
// and follow the naming convention: SLEEPSHIP_<SECTION>_<KEY>
//
// Supported environment variables:
// - SLEEPSHIP_PROJECT_DIR: Project directory
// - SLEEPSHIP_SYNC_DEFAULT_TASK_FILE: Default task file
// - SLEEPSHIP_SYNC_MAX_RETRIES: Maximum number of retries
// - SLEEPSHIP_SYNC_LOG_DIR: Log directory
// - SLEEPSHIP_SYNC_START_FROM: Start from specified task number
// - SLEEPSHIP_CLAUDE_FLAGS: Claude Code flags (comma-separated)
func LoadFromEnv() *EnvConfig {
	cfg := &EnvConfig{
		MaxRetries: -1, // Use -1 to indicate not set
		StartFrom:  -1, // Use -1 to indicate not set
	}

	// Project directory
	if val := os.Getenv("SLEEPSHIP_PROJECT_DIR"); val != "" {
		cfg.ProjectDir = val
	}

	// Default task file
	if val := os.Getenv("SLEEPSHIP_SYNC_DEFAULT_TASK_FILE"); val != "" {
		cfg.DefaultTaskFile = val
	}

	// Max retries
	if val := os.Getenv("SLEEPSHIP_SYNC_MAX_RETRIES"); val != "" {
		if n, err := strconv.Atoi(val); err == nil && n >= 0 {
			cfg.MaxRetries = n
		}
	}

	// Log directory
	if val := os.Getenv("SLEEPSHIP_SYNC_LOG_DIR"); val != "" {
		cfg.LogDir = val
	}

	// Start from
	if val := os.Getenv("SLEEPSHIP_SYNC_START_FROM"); val != "" {
		if n, err := strconv.Atoi(val); err == nil && n >= 1 {
			cfg.StartFrom = n
		}
	}

	// Claude flags (comma-separated)
	if val := os.Getenv("SLEEPSHIP_CLAUDE_FLAGS"); val != "" {
		flags := strings.Split(val, ",")
		for i, flag := range flags {
			flags[i] = strings.TrimSpace(flag)
		}
		cfg.ClaudeFlags = flags
	}

	return cfg
}

// HasProjectDir checks if ProjectDir has been set via environment variable.
func (c *EnvConfig) HasProjectDir() bool {
	return c.ProjectDir != ""
}

// HasDefaultTaskFile checks if DefaultTaskFile has been set via environment variable.
func (c *EnvConfig) HasDefaultTaskFile() bool {
	return c.DefaultTaskFile != ""
}

// HasMaxRetries checks if MaxRetries has been set via environment variable.
func (c *EnvConfig) HasMaxRetries() bool {
	return c.MaxRetries >= 0
}

// HasLogDir checks if LogDir has been set via environment variable.
func (c *EnvConfig) HasLogDir() bool {
	return c.LogDir != ""
}

// HasStartFrom checks if StartFrom has been set via environment variable.
func (c *EnvConfig) HasStartFrom() bool {
	return c.StartFrom >= 1
}

// HasClaudeFlags checks if ClaudeFlags have been set via environment variable.
func (c *EnvConfig) HasClaudeFlags() bool {
	return len(c.ClaudeFlags) > 0
}
