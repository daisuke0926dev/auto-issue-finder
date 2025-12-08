package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveAlias(t *testing.T) {
	tests := []struct {
		name      string
		alias     string
		aliases   map[string]string
		want      string
		wantError bool
	}{
		{
			name:    "simple alias",
			alias:   "dev",
			aliases: map[string]string{"dev": "sync tasks-dev.txt"},
			want:    "sync tasks-dev.txt",
		},
		{
			name:    "nested alias",
			alias:   "quick",
			aliases: map[string]string{"dev": "sync tasks-dev.txt", "quick": "dev --max-retries=1"},
			want:    "sync tasks-dev.txt --max-retries=1",
		},
		{
			name:      "circular reference",
			alias:     "a",
			aliases:   map[string]string{"a": "b", "b": "a"},
			wantError: true,
		},
		{
			name:      "self reference",
			alias:     "loop",
			aliases:   map[string]string{"loop": "loop"},
			wantError: true,
		},
		{
			name:      "non-existent alias",
			alias:     "missing",
			aliases:   map[string]string{"dev": "sync tasks-dev.txt"},
			wantError: true,
		},
		{
			name:    "three level nesting",
			alias:   "fast",
			aliases: map[string]string{"base": "sync tasks.txt", "dev": "base --max-retries=3", "fast": "dev --start-from=2"},
			want:    "sync tasks.txt --max-retries=3 --start-from=2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveAlias(tt.alias, tt.aliases)
			if tt.wantError {
				if err == nil {
					t.Errorf("ResolveAlias() expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("ResolveAlias() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("ResolveAlias() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExpandAliasArgs(t *testing.T) {
	tests := []struct {
		name    string
		command string
		args    []string
		want    string
	}{
		{
			name:    "no args",
			command: "sync tasks-dev.txt",
			args:    []string{},
			want:    "sync tasks-dev.txt",
		},
		{
			name:    "with args",
			command: "sync tasks-dev.txt",
			args:    []string{"--max-retries=5"},
			want:    "sync tasks-dev.txt --max-retries=5",
		},
		{
			name:    "multiple args",
			command: "sync tasks-dev.txt",
			args:    []string{"--max-retries=5", "--start-from=2"},
			want:    "sync tasks-dev.txt --max-retries=5 --start-from=2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExpandAliasArgs(tt.command, tt.args)
			if got != tt.want {
				t.Errorf("ExpandAliasArgs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoadAliases(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()

	// Test 1: No config file
	t.Run("no config file", func(t *testing.T) {
		// Change to temp directory
		oldDir, _ := os.Getwd()
		defer func() { _ = os.Chdir(oldDir) }()
		_ = os.Chdir(tmpDir)

		aliases, err := LoadAliases()
		if err != nil {
			t.Errorf("LoadAliases() error = %v, want nil", err)
		}
		if len(aliases) != 0 {
			t.Errorf("LoadAliases() returned %d aliases, want 0", len(aliases))
		}
	})

	// Test 2: Valid config file
	t.Run("valid config file", func(t *testing.T) {
		configContent := `[aliases]
dev = "sync tasks-dev.txt"
test = "sync tasks-test.txt --max-retries=5"
prod = "sync tasks-prod.txt --max-retries=10"
`
		configPath := filepath.Join(tmpDir, ".sleepship.toml")
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatal(err)
		}

		// Change to temp directory
		oldDir, _ := os.Getwd()
		defer func() { _ = os.Chdir(oldDir) }()
		_ = os.Chdir(tmpDir)

		aliases, err := LoadAliases()
		if err != nil {
			t.Errorf("LoadAliases() error = %v, want nil", err)
		}
		if len(aliases) != 3 {
			t.Errorf("LoadAliases() returned %d aliases, want 3", len(aliases))
		}

		expectedAliases := map[string]string{
			"dev":  "sync tasks-dev.txt",
			"test": "sync tasks-test.txt --max-retries=5",
			"prod": "sync tasks-prod.txt --max-retries=10",
		}

		for name, cmd := range expectedAliases {
			if aliases[name] != cmd {
				t.Errorf("alias %s = %v, want %v", name, aliases[name], cmd)
			}
		}
	})

	// Test 3: Invalid config file
	t.Run("invalid config file", func(t *testing.T) {
		invalidConfigPath := filepath.Join(tmpDir, ".sleepship-invalid.toml")
		if err := os.WriteFile(invalidConfigPath, []byte("invalid toml [[["), 0644); err != nil {
			t.Fatal(err)
		}

		// Create a subdirectory and change to it
		subDir := filepath.Join(tmpDir, "invalid-test")
		_ = os.Mkdir(subDir, 0755)
		oldDir, _ := os.Getwd()
		defer func() { _ = os.Chdir(oldDir) }()
		_ = os.Chdir(subDir)

		// Copy invalid config to current directory
		_ = os.WriteFile(".sleepship.toml", []byte("invalid toml [[["), 0644)

		_, err := LoadAliases()
		if err == nil {
			t.Error("LoadAliases() expected error for invalid config, got nil")
		}
	})
}
