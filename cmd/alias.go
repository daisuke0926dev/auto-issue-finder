package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/isiidaisuke0926/sleepship/internal/config"
	"github.com/spf13/cobra"
)

var aliasCmd = &cobra.Command{
	Use:   "alias",
	Short: "Manage command aliases",
	Long: `Manage command aliases defined in .sleepship.toml.

Aliases allow you to define shortcuts for frequently used commands.

Example .sleepship.toml:
  [aliases]
  dev = "sync tasks-dev.txt"
  test = "sync tasks-test.txt --max-retries=5"
  prod = "sync tasks-prod.txt --max-retries=10"

Then you can use:
  sleepship dev
  sleepship test
  sleepship prod`,
}

var aliasListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all aliases",
	Long:  "List all command aliases defined in .sleepship.toml",
	RunE:  runAliasList,
}

var aliasGetCmd = &cobra.Command{
	Use:   "get <name>",
	Short: "Get a specific alias",
	Long:  "Display the command for a specific alias",
	Args:  cobra.ExactArgs(1),
	RunE:  runAliasGet,
}

func init() {
	rootCmd.AddCommand(aliasCmd)
	aliasCmd.AddCommand(aliasListCmd)
	aliasCmd.AddCommand(aliasGetCmd)
}

func runAliasList(cmd *cobra.Command, args []string) error {
	aliases, err := config.LoadAliases()
	if err != nil {
		return fmt.Errorf("failed to load aliases: %w", err)
	}

	if len(aliases) == 0 {
		// Check if config file exists
		configPath := findConfigPath()
		if configPath == "" {
			fmt.Println("No .sleepship.toml file found.")
			fmt.Println("\nCreate a .sleepship.toml file in your project directory or home directory with:")
			fmt.Println("\n[aliases]")
			fmt.Println("dev = \"sync tasks-dev.txt\"")
			fmt.Println("test = \"sync tasks-test.txt --max-retries=5\"")
			return nil
		}

		fmt.Println("No aliases defined in .sleepship.toml")
		return nil
	}

	fmt.Printf("Aliases (%d):\n\n", len(aliases))

	// Sort aliases by name
	names := make([]string, 0, len(aliases))
	for name := range aliases {
		names = append(names, name)
	}
	sort.Strings(names)

	// Find max name length for alignment
	maxLen := 0
	for _, name := range names {
		if len(name) > maxLen {
			maxLen = len(name)
		}
	}

	// Display aliases
	for _, name := range names {
		resolved, err := config.ResolveAlias(name, aliases)
		if err != nil {
			fmt.Printf("  %-*s -> %s (error: %v)\n", maxLen, name, aliases[name], err)
		} else if resolved != aliases[name] {
			// Show both original and resolved if different
			fmt.Printf("  %-*s -> %s (resolves to: %s)\n", maxLen, name, aliases[name], resolved)
		} else {
			fmt.Printf("  %-*s -> %s\n", maxLen, name, aliases[name])
		}
	}

	fmt.Printf("\nConfig file: %s\n", findConfigPath())

	return nil
}

func runAliasGet(cmd *cobra.Command, args []string) error {
	aliasName := args[0]

	aliases, err := config.LoadAliases()
	if err != nil {
		return fmt.Errorf("failed to load aliases: %w", err)
	}

	command, exists := aliases[aliasName]
	if !exists {
		return fmt.Errorf("alias not found: %s", aliasName)
	}

	fmt.Printf("Alias: %s\n", aliasName)
	fmt.Printf("Command: %s\n", command)

	// Show resolved command if different
	resolved, err := config.ResolveAlias(aliasName, aliases)
	if err != nil {
		fmt.Printf("Error resolving: %v\n", err)
	} else if resolved != command {
		fmt.Printf("Resolves to: %s\n", resolved)
	}

	return nil
}

func findConfigPath() string {
	// Check current directory
	cwd, err := os.Getwd()
	if err == nil {
		configPath := filepath.Join(cwd, ".sleepship.toml")
		if _, err := os.Stat(configPath); err == nil {
			return configPath
		}
	}

	// Check home directory
	homeDir, err := os.UserHomeDir()
	if err == nil {
		configPath := filepath.Join(homeDir, ".sleepship.toml")
		if _, err := os.Stat(configPath); err == nil {
			return configPath
		}
	}

	return ""
}
