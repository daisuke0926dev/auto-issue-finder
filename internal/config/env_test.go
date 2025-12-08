package config

import (
	"os"
	"testing"
)

func TestLoadFromEnv(t *testing.T) {
	// Save original env vars
	originalVars := map[string]string{
		"SLEEPSHIP_PROJECT_DIR":            os.Getenv("SLEEPSHIP_PROJECT_DIR"),
		"SLEEPSHIP_SYNC_DEFAULT_TASK_FILE": os.Getenv("SLEEPSHIP_SYNC_DEFAULT_TASK_FILE"),
		"SLEEPSHIP_SYNC_MAX_RETRIES":       os.Getenv("SLEEPSHIP_SYNC_MAX_RETRIES"),
		"SLEEPSHIP_SYNC_LOG_DIR":           os.Getenv("SLEEPSHIP_SYNC_LOG_DIR"),
		"SLEEPSHIP_SYNC_START_FROM":        os.Getenv("SLEEPSHIP_SYNC_START_FROM"),
		"SLEEPSHIP_CLAUDE_FLAGS":           os.Getenv("SLEEPSHIP_CLAUDE_FLAGS"),
	}

	// Restore original env vars after test
	defer func() {
		for key, val := range originalVars {
			if val == "" {
				_ = os.Unsetenv(key)
			} else {
				_ = os.Setenv(key, val)
			}
		}
	}()

	tests := []struct {
		name     string
		envVars  map[string]string
		expected *EnvConfig
	}{
		{
			name: "all environment variables set",
			envVars: map[string]string{
				"SLEEPSHIP_PROJECT_DIR":            "/test/project",
				"SLEEPSHIP_SYNC_DEFAULT_TASK_FILE": "tasks-default.txt",
				"SLEEPSHIP_SYNC_MAX_RETRIES":       "5",
				"SLEEPSHIP_SYNC_LOG_DIR":           "/test/logs",
				"SLEEPSHIP_SYNC_START_FROM":        "2",
				"SLEEPSHIP_CLAUDE_FLAGS":           "--flag1, --flag2",
			},
			expected: &EnvConfig{
				ProjectDir:      "/test/project",
				DefaultTaskFile: "tasks-default.txt",
				MaxRetries:      5,
				LogDir:          "/test/logs",
				StartFrom:       2,
				ClaudeFlags:     []string{"--flag1", "--flag2"},
			},
		},
		{
			name:    "no environment variables set",
			envVars: map[string]string{},
			expected: &EnvConfig{
				MaxRetries: -1,
				StartFrom:  -1,
			},
		},
		{
			name: "invalid max retries",
			envVars: map[string]string{
				"SLEEPSHIP_SYNC_MAX_RETRIES": "invalid",
			},
			expected: &EnvConfig{
				MaxRetries: -1,
				StartFrom:  -1,
			},
		},
		{
			name: "negative max retries",
			envVars: map[string]string{
				"SLEEPSHIP_SYNC_MAX_RETRIES": "-1",
			},
			expected: &EnvConfig{
				MaxRetries: -1,
				StartFrom:  -1,
			},
		},
		{
			name: "invalid start from",
			envVars: map[string]string{
				"SLEEPSHIP_SYNC_START_FROM": "invalid",
			},
			expected: &EnvConfig{
				MaxRetries: -1,
				StartFrom:  -1,
			},
		},
		{
			name: "zero start from",
			envVars: map[string]string{
				"SLEEPSHIP_SYNC_START_FROM": "0",
			},
			expected: &EnvConfig{
				MaxRetries: -1,
				StartFrom:  -1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all env vars first
			_ = os.Unsetenv("SLEEPSHIP_PROJECT_DIR")
			_ = os.Unsetenv("SLEEPSHIP_SYNC_DEFAULT_TASK_FILE")
			_ = os.Unsetenv("SLEEPSHIP_SYNC_MAX_RETRIES")
			_ = os.Unsetenv("SLEEPSHIP_SYNC_LOG_DIR")
			_ = os.Unsetenv("SLEEPSHIP_SYNC_START_FROM")
			_ = os.Unsetenv("SLEEPSHIP_CLAUDE_FLAGS")

			// Set test env vars
			for key, val := range tt.envVars {
				_ = os.Setenv(key, val)
			}

			// Load config
			cfg := LoadFromEnv()

			// Verify
			if cfg.ProjectDir != tt.expected.ProjectDir {
				t.Errorf("ProjectDir = %v, want %v", cfg.ProjectDir, tt.expected.ProjectDir)
			}
			if cfg.DefaultTaskFile != tt.expected.DefaultTaskFile {
				t.Errorf("DefaultTaskFile = %v, want %v", cfg.DefaultTaskFile, tt.expected.DefaultTaskFile)
			}
			if cfg.MaxRetries != tt.expected.MaxRetries {
				t.Errorf("MaxRetries = %v, want %v", cfg.MaxRetries, tt.expected.MaxRetries)
			}
			if cfg.LogDir != tt.expected.LogDir {
				t.Errorf("LogDir = %v, want %v", cfg.LogDir, tt.expected.LogDir)
			}
			if cfg.StartFrom != tt.expected.StartFrom {
				t.Errorf("StartFrom = %v, want %v", cfg.StartFrom, tt.expected.StartFrom)
			}
			if len(cfg.ClaudeFlags) != len(tt.expected.ClaudeFlags) {
				t.Errorf("ClaudeFlags length = %v, want %v", len(cfg.ClaudeFlags), len(tt.expected.ClaudeFlags))
			} else {
				for i, flag := range cfg.ClaudeFlags {
					if flag != tt.expected.ClaudeFlags[i] {
						t.Errorf("ClaudeFlags[%d] = %v, want %v", i, flag, tt.expected.ClaudeFlags[i])
					}
				}
			}
		})
	}
}

func TestEnvConfigHasValue(t *testing.T) {
	tests := []struct {
		name   string
		config *EnvConfig
		checks map[string]bool
	}{
		{
			name: "all values set",
			config: &EnvConfig{
				ProjectDir:      "/test",
				DefaultTaskFile: "tasks.txt",
				MaxRetries:      5,
				LogDir:          "logs",
				StartFrom:       2,
				ClaudeFlags:     []string{"--flag"},
			},
			checks: map[string]bool{
				"ProjectDir":      true,
				"DefaultTaskFile": true,
				"MaxRetries":      true,
				"LogDir":          true,
				"StartFrom":       true,
				"ClaudeFlags":     true,
			},
		},
		{
			name: "no values set",
			config: &EnvConfig{
				MaxRetries: -1,
				StartFrom:  -1,
			},
			checks: map[string]bool{
				"ProjectDir":      false,
				"DefaultTaskFile": false,
				"MaxRetries":      false,
				"LogDir":          false,
				"StartFrom":       false,
				"ClaudeFlags":     false,
			},
		},
		{
			name: "zero max retries is valid",
			config: &EnvConfig{
				MaxRetries: 0,
				StartFrom:  -1,
			},
			checks: map[string]bool{
				"MaxRetries": true,
				"StartFrom":  false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if hasProjectDir, expected := tt.config.HasProjectDir(), tt.checks["ProjectDir"]; hasProjectDir != expected {
				t.Errorf("HasProjectDir() = %v, want %v", hasProjectDir, expected)
			}
			if hasDefaultTaskFile, expected := tt.config.HasDefaultTaskFile(), tt.checks["DefaultTaskFile"]; hasDefaultTaskFile != expected {
				t.Errorf("HasDefaultTaskFile() = %v, want %v", hasDefaultTaskFile, expected)
			}
			if hasMaxRetries, expected := tt.config.HasMaxRetries(), tt.checks["MaxRetries"]; hasMaxRetries != expected {
				t.Errorf("HasMaxRetries() = %v, want %v", hasMaxRetries, expected)
			}
			if hasLogDir, expected := tt.config.HasLogDir(), tt.checks["LogDir"]; hasLogDir != expected {
				t.Errorf("HasLogDir() = %v, want %v", hasLogDir, expected)
			}
			if hasStartFrom, expected := tt.config.HasStartFrom(), tt.checks["StartFrom"]; hasStartFrom != expected {
				t.Errorf("HasStartFrom() = %v, want %v", hasStartFrom, expected)
			}
			if hasClaudeFlags, expected := tt.config.HasClaudeFlags(), tt.checks["ClaudeFlags"]; hasClaudeFlags != expected {
				t.Errorf("HasClaudeFlags() = %v, want %v", hasClaudeFlags, expected)
			}
		})
	}
}

func TestMergeConfig(t *testing.T) {
	tests := []struct {
		name          string
		cliConfig     *Config
		envConfig     *Config
		defaultConfig *Config
		expected      *Config
	}{
		{
			name: "CLI takes precedence over env and default",
			cliConfig: &Config{
				ProjectDir: "/cli/project",
				MaxRetries: 10,
				LogDir:     "cli-logs",
				StartFrom:  3,
			},
			envConfig: &Config{
				ProjectDir: "/env/project",
				MaxRetries: 5,
				LogDir:     "env-logs",
				StartFrom:  2,
			},
			defaultConfig: &Config{
				ProjectDir: "/default/project",
				MaxRetries: 3,
				LogDir:     "logs",
				StartFrom:  1,
			},
			expected: &Config{
				ProjectDir: "/cli/project",
				MaxRetries: 10,
				LogDir:     "cli-logs",
				StartFrom:  3,
			},
		},
		{
			name: "Env takes precedence over default",
			cliConfig: &Config{
				MaxRetries: -1,
				StartFrom:  -1,
			},
			envConfig: &Config{
				ProjectDir: "/env/project",
				MaxRetries: 5,
				LogDir:     "env-logs",
				StartFrom:  2,
			},
			defaultConfig: &Config{
				ProjectDir: "/default/project",
				MaxRetries: 3,
				LogDir:     "logs",
				StartFrom:  1,
			},
			expected: &Config{
				ProjectDir: "/env/project",
				MaxRetries: 5,
				LogDir:     "env-logs",
				StartFrom:  2,
			},
		},
		{
			name: "Falls back to default",
			cliConfig: &Config{
				MaxRetries: -1,
				StartFrom:  -1,
			},
			envConfig: &Config{
				MaxRetries: -1,
				StartFrom:  -1,
			},
			defaultConfig: &Config{
				ProjectDir: "/default/project",
				MaxRetries: 3,
				LogDir:     "logs",
				StartFrom:  1,
			},
			expected: &Config{
				ProjectDir: "/default/project",
				MaxRetries: 3,
				LogDir:     "logs",
				StartFrom:  1,
			},
		},
		{
			name: "Partial overrides",
			cliConfig: &Config{
				ProjectDir: "/cli/project",
				MaxRetries: -1,
				StartFrom:  -1,
			},
			envConfig: &Config{
				MaxRetries: 7,
				LogDir:     "env-logs",
				StartFrom:  -1,
			},
			defaultConfig: &Config{
				ProjectDir: "/default/project",
				MaxRetries: 3,
				LogDir:     "logs",
				StartFrom:  1,
			},
			expected: &Config{
				ProjectDir: "/cli/project",
				MaxRetries: 7,
				LogDir:     "env-logs",
				StartFrom:  1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			merged := MergeConfig(tt.cliConfig, tt.envConfig, tt.defaultConfig)

			if merged.ProjectDir != tt.expected.ProjectDir {
				t.Errorf("ProjectDir = %v, want %v", merged.ProjectDir, tt.expected.ProjectDir)
			}
			if merged.MaxRetries != tt.expected.MaxRetries {
				t.Errorf("MaxRetries = %v, want %v", merged.MaxRetries, tt.expected.MaxRetries)
			}
			if merged.LogDir != tt.expected.LogDir {
				t.Errorf("LogDir = %v, want %v", merged.LogDir, tt.expected.LogDir)
			}
			if merged.StartFrom != tt.expected.StartFrom {
				t.Errorf("StartFrom = %v, want %v", merged.StartFrom, tt.expected.StartFrom)
			}
		})
	}
}

func TestFromEnv(t *testing.T) {
	envCfg := &EnvConfig{
		ProjectDir:      "/test/project",
		DefaultTaskFile: "tasks.txt",
		MaxRetries:      5,
		LogDir:          "test-logs",
		StartFrom:       2,
		ClaudeFlags:     []string{"--flag1", "--flag2"},
	}

	cfg := FromEnv(envCfg)

	if cfg.ProjectDir != envCfg.ProjectDir {
		t.Errorf("ProjectDir = %v, want %v", cfg.ProjectDir, envCfg.ProjectDir)
	}
	if cfg.DefaultTaskFile != envCfg.DefaultTaskFile {
		t.Errorf("DefaultTaskFile = %v, want %v", cfg.DefaultTaskFile, envCfg.DefaultTaskFile)
	}
	if cfg.MaxRetries != envCfg.MaxRetries {
		t.Errorf("MaxRetries = %v, want %v", cfg.MaxRetries, envCfg.MaxRetries)
	}
	if cfg.LogDir != envCfg.LogDir {
		t.Errorf("LogDir = %v, want %v", cfg.LogDir, envCfg.LogDir)
	}
	if cfg.StartFrom != envCfg.StartFrom {
		t.Errorf("StartFrom = %v, want %v", cfg.StartFrom, envCfg.StartFrom)
	}
	if len(cfg.ClaudeFlags) != len(envCfg.ClaudeFlags) {
		t.Errorf("ClaudeFlags length = %v, want %v", len(cfg.ClaudeFlags), len(envCfg.ClaudeFlags))
	}
}
