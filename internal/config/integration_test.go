package config

import (
	"os"
	"testing"
)

// TestConfigPriority tests the configuration priority order:
// CLI > Environment > Default
func TestConfigPriority(t *testing.T) {
	// Save and restore environment
	originalMaxRetries := os.Getenv("SLEEPSHIP_SYNC_MAX_RETRIES")
	defer func() {
		if originalMaxRetries == "" {
			_ = os.Unsetenv("SLEEPSHIP_SYNC_MAX_RETRIES")
		} else {
			_ = os.Setenv("SLEEPSHIP_SYNC_MAX_RETRIES", originalMaxRetries)
		}
	}()

	// Set environment variable
	_ = os.Setenv("SLEEPSHIP_SYNC_MAX_RETRIES", "7")

	defaultConfig := NewDefaultConfig()
	envConfig := LoadFromEnv()
	cliConfig := &Config{
		MaxRetries: 10,
	}

	// Test: CLI should take precedence
	merged := MergeConfig(cliConfig, FromEnv(envConfig), defaultConfig)
	if merged.MaxRetries != 10 {
		t.Errorf("CLI priority failed: got %d, want 10", merged.MaxRetries)
	}

	// Test: Environment should take precedence over default
	cliConfig.MaxRetries = -1 // Not set
	merged = MergeConfig(cliConfig, FromEnv(envConfig), defaultConfig)
	if merged.MaxRetries != 7 {
		t.Errorf("Env priority failed: got %d, want 7", merged.MaxRetries)
	}

	// Test: Default should be used when nothing else is set
	_ = os.Unsetenv("SLEEPSHIP_SYNC_MAX_RETRIES")
	envConfig = LoadFromEnv()
	merged = MergeConfig(cliConfig, FromEnv(envConfig), defaultConfig)
	if merged.MaxRetries != 3 {
		t.Errorf("Default priority failed: got %d, want 3", merged.MaxRetries)
	}
}

func TestConfigMergeAllFields(t *testing.T) {
	// Save and restore all environment variables
	envVars := map[string]string{
		"SLEEPSHIP_PROJECT_DIR":            os.Getenv("SLEEPSHIP_PROJECT_DIR"),
		"SLEEPSHIP_SYNC_DEFAULT_TASK_FILE": os.Getenv("SLEEPSHIP_SYNC_DEFAULT_TASK_FILE"),
		"SLEEPSHIP_SYNC_MAX_RETRIES":       os.Getenv("SLEEPSHIP_SYNC_MAX_RETRIES"),
		"SLEEPSHIP_SYNC_LOG_DIR":           os.Getenv("SLEEPSHIP_SYNC_LOG_DIR"),
		"SLEEPSHIP_SYNC_START_FROM":        os.Getenv("SLEEPSHIP_SYNC_START_FROM"),
		"SLEEPSHIP_CLAUDE_FLAGS":           os.Getenv("SLEEPSHIP_CLAUDE_FLAGS"),
	}
	defer func() {
		for key, val := range envVars {
			if val == "" {
				_ = os.Unsetenv(key)
			} else {
				_ = os.Setenv(key, val)
			}
		}
	}()

	// Set all environment variables
	_ = os.Setenv("SLEEPSHIP_PROJECT_DIR", "/env/project")
	_ = os.Setenv("SLEEPSHIP_SYNC_DEFAULT_TASK_FILE", "env-tasks.txt")
	_ = os.Setenv("SLEEPSHIP_SYNC_MAX_RETRIES", "5")
	_ = os.Setenv("SLEEPSHIP_SYNC_LOG_DIR", "env-logs")
	_ = os.Setenv("SLEEPSHIP_SYNC_START_FROM", "2")
	_ = os.Setenv("SLEEPSHIP_CLAUDE_FLAGS", "--env-flag1,--env-flag2")

	defaultConfig := NewDefaultConfig()
	envConfig := LoadFromEnv()
	cliConfig := &Config{
		ProjectDir: "/cli/project",
		MaxRetries: -1, // Not set
		StartFrom:  -1, // Not set
	}

	merged := MergeConfig(cliConfig, FromEnv(envConfig), defaultConfig)

	// CLI takes precedence for ProjectDir
	if merged.ProjectDir != "/cli/project" {
		t.Errorf("ProjectDir = %v, want /cli/project", merged.ProjectDir)
	}

	// Environment takes precedence for fields not set by CLI
	if merged.DefaultTaskFile != "env-tasks.txt" {
		t.Errorf("DefaultTaskFile = %v, want env-tasks.txt", merged.DefaultTaskFile)
	}
	if merged.MaxRetries != 5 {
		t.Errorf("MaxRetries = %v, want 5", merged.MaxRetries)
	}
	if merged.LogDir != "env-logs" {
		t.Errorf("LogDir = %v, want env-logs", merged.LogDir)
	}
	if merged.StartFrom != 2 {
		t.Errorf("StartFrom = %v, want 2", merged.StartFrom)
	}
	if len(merged.ClaudeFlags) != 2 {
		t.Errorf("ClaudeFlags length = %v, want 2", len(merged.ClaudeFlags))
	}
}
