//go:build ignore

package main

import (
	"fmt"
	"os"

	"github.com/isiidaisuke0926/sleepship/internal/config"
)

func main() {
	fmt.Println("Testing environment variable loading...")
	fmt.Println()

	// Print current environment variables
	envVars := []string{
		"SLEEPSHIP_PROJECT_DIR",
		"SLEEPSHIP_SYNC_DEFAULT_TASK_FILE",
		"SLEEPSHIP_SYNC_MAX_RETRIES",
		"SLEEPSHIP_SYNC_LOG_DIR",
		"SLEEPSHIP_SYNC_START_FROM",
		"SLEEPSHIP_CLAUDE_FLAGS",
	}

	fmt.Println("Current environment variables:")
	for _, key := range envVars {
		val := os.Getenv(key)
		if val != "" {
			fmt.Printf("  %s = %s\n", key, val)
		} else {
			fmt.Printf("  %s = (not set)\n", key)
		}
	}
	fmt.Println()

	// Load configuration
	envConfig := config.LoadFromEnv()

	fmt.Println("Loaded configuration:")
	fmt.Printf("  ProjectDir: %s (has value: %v)\n", envConfig.ProjectDir, envConfig.HasProjectDir())
	fmt.Printf("  DefaultTaskFile: %s (has value: %v)\n", envConfig.DefaultTaskFile, envConfig.HasDefaultTaskFile())
	fmt.Printf("  MaxRetries: %d (has value: %v)\n", envConfig.MaxRetries, envConfig.HasMaxRetries())
	fmt.Printf("  LogDir: %s (has value: %v)\n", envConfig.LogDir, envConfig.HasLogDir())
	fmt.Printf("  StartFrom: %d (has value: %v)\n", envConfig.StartFrom, envConfig.HasStartFrom())
	fmt.Printf("  ClaudeFlags: %v (has value: %v)\n", envConfig.ClaudeFlags, envConfig.HasClaudeFlags())
	fmt.Println()

	// Test merging
	defaultConfig := config.NewDefaultConfig()
	cliConfig := &config.Config{
		MaxRetries: -1,
		StartFrom:  -1,
	}

	mergedConfig := config.MergeConfig(cliConfig, config.FromEnv(envConfig), defaultConfig)

	fmt.Println("Merged configuration:")
	fmt.Printf("  ProjectDir: %s\n", mergedConfig.ProjectDir)
	fmt.Printf("  DefaultTaskFile: %s\n", mergedConfig.DefaultTaskFile)
	fmt.Printf("  MaxRetries: %d\n", mergedConfig.MaxRetries)
	fmt.Printf("  LogDir: %s\n", mergedConfig.LogDir)
	fmt.Printf("  StartFrom: %d\n", mergedConfig.StartFrom)
	fmt.Printf("  ClaudeFlags: %v\n", mergedConfig.ClaudeFlags)
}
