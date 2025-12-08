package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

// AliasConfig represents alias configuration from .sleepship.toml
type AliasConfig struct {
	Aliases map[string]string `toml:"aliases"`
}

// LoadAliases loads alias configuration from .sleepship.toml
// It searches for .sleepship.toml in the current directory first,
// then in the home directory.
func LoadAliases() (map[string]string, error) {
	// Search for config file in current directory
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	configPath := filepath.Join(cwd, ".sleepship.toml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Try home directory
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		configPath = filepath.Join(homeDir, ".sleepship.toml")
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			// No config file found, return empty map
			return make(map[string]string), nil
		}
	}

	// Load config file
	var config AliasConfig
	if _, err := toml.DecodeFile(configPath, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return config.Aliases, nil
}

// ResolveAlias resolves an alias to its command
// It detects and prevents circular references
func ResolveAlias(alias string, aliases map[string]string) (string, error) {
	visited := make(map[string]bool)
	return resolveAliasRecursive(alias, aliases, visited)
}

func resolveAliasRecursive(alias string, aliases map[string]string, visited map[string]bool) (string, error) {
	// Check for circular reference before processing
	if visited[alias] {
		return "", fmt.Errorf("circular reference detected in alias: %s", alias)
	}

	// Check if alias exists
	command, exists := aliases[alias]
	if !exists {
		return "", fmt.Errorf("alias not found: %s", alias)
	}

	// Mark as visited
	visited[alias] = true

	// Parse command to check if it contains another alias
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return command, nil
	}

	// Check if the first word is an alias
	firstWord := parts[0]
	if _, isAlias := aliases[firstWord]; isAlias {
		// Recursively resolve the alias with the same visited map
		resolvedFirst, err := resolveAliasRecursive(firstWord, aliases, visited)
		if err != nil {
			return "", err
		}
		// Replace the first word with resolved alias
		parts[0] = resolvedFirst
		return strings.Join(parts, " "), nil
	}

	return command, nil
}

// ExpandAliasArgs expands an alias with its arguments
// Example: alias="sync tasks-dev.txt", args=["--max-retries=5"]
// Result: "sync tasks-dev.txt --max-retries=5"
func ExpandAliasArgs(command string, args []string) string {
	if len(args) == 0 {
		return command
	}
	return command + " " + strings.Join(args, " ")
}
