package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "auto-issue-finder",
	Short: "Analyze GitHub issues and discover patterns",
	Long: `auto-issue-finder is a CLI tool that automatically analyzes GitHub repository issues,
identifies patterns, and provides actionable recommendations for improvement.`,
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}
