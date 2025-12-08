package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/isiidaisuke0926/sleepship/internal/config"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "sleepship",
	Short: "Autonomous development system with Claude Code",
	Long: `sleepship is a CLI tool for autonomous software development.
It executes development tasks synchronously using Claude Code, with automatic
error detection and correction.`,
}

// Execute runs the root command
func Execute() {
	// Pre-process arguments to resolve aliases
	if len(os.Args) > 1 {
		firstArg := os.Args[1]

		// Skip alias resolution for built-in commands
		if firstArg == "alias" || firstArg == "help" || firstArg == "--help" || firstArg == "-h" {
			// Execute normally
			if err := rootCmd.Execute(); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			return
		}

		// Try to resolve as alias
		aliases, err := config.LoadAliases()
		if err == nil && len(aliases) > 0 {
			if _, exists := aliases[firstArg]; exists {
				// Resolve the alias
				resolved, err := config.ResolveAlias(firstArg, aliases)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error resolving alias '%s': %v\n", firstArg, err)
					os.Exit(1)
				}

				// Expand with remaining arguments
				remainingArgs := os.Args[2:]
				fullCommand := config.ExpandAliasArgs(resolved, remainingArgs)

				// Parse the expanded command
				parts := strings.Fields(fullCommand)
				if len(parts) == 0 {
					fmt.Fprintln(os.Stderr, "Error: empty command after alias expansion")
					os.Exit(1)
				}

				// Update os.Args with resolved command
				os.Args = append([]string{os.Args[0]}, parts...)
			}
		}
	}

	// Execute the command (either original or resolved)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}
