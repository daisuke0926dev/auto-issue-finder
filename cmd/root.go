package cmd

import (
	"fmt"
	"os"

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
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}
