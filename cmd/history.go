package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/isiidaisuke0926/sleepship/internal/history"
	"github.com/spf13/cobra"
)

var (
	historyLast   int
	historyFailed bool
)

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "Show task execution history",
	Long: `Show the history of task executions.

Examples:
  sleepship history                 # Show all history
  sleepship history --last 5        # Show last 5 executions
  sleepship history --last 1        # Show last execution
  sleepship history --failed        # Show only failed executions`,
	RunE: runHistory,
}

func init() {
	rootCmd.AddCommand(historyCmd)

	historyCmd.Flags().IntVar(&historyLast, "last", 0, "Show last N executions (0 = all)")
	historyCmd.Flags().BoolVar(&historyFailed, "failed", false, "Show only failed executions")
}

func runHistory(cmd *cobra.Command, args []string) error {
	// Get project directory
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Load history
	hist, err := history.Load(dir)
	if err != nil {
		return fmt.Errorf("failed to load history: %w", err)
	}

	if len(hist.Entries) == 0 {
		fmt.Println("📋 No execution history found")
		return nil
	}

	// Filter entries
	entries := hist.Entries
	if historyFailed {
		entries = hist.GetFailed()
		if len(entries) == 0 {
			fmt.Println("✅ No failed executions found")
			return nil
		}
	} else if historyLast > 0 {
		entries = hist.GetLast(historyLast)
	}

	// Display entries
	displayHistory(entries)

	return nil
}

func displayHistory(entries []history.Entry) {
	// Create color printers
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()

	fmt.Printf("📋 Task Execution History (%d entries)\n\n", len(entries))

	// Calculate column widths
	maxTaskFileLen := 0
	for _, entry := range entries {
		taskFile := filepath.Base(entry.TaskFile)
		if len(taskFile) > maxTaskFileLen {
			maxTaskFileLen = len(taskFile)
		}
	}
	if maxTaskFileLen < 20 {
		maxTaskFileLen = 20
	}
	if maxTaskFileLen > 50 {
		maxTaskFileLen = 50
	}

	// Header
	headerFormat := "%-6s %-" + fmt.Sprintf("%d", maxTaskFileLen) + "s %-20s %-10s %-6s %-8s %s\n"
	fmt.Printf(headerFormat, "Status", "Task File", "Executed At", "Duration", "Tasks", "Retries", "Branch")
	fmt.Println(strings.Repeat("-", maxTaskFileLen+80))

	// Entries
	entryFormat := "%-6s %-" + fmt.Sprintf("%d", maxTaskFileLen) + "s %-20s %-10s %-6d %-8d %s\n"
	for _, entry := range entries {
		// Status
		var status string
		if entry.Success {
			status = green("✅")
		} else {
			status = red("❌")
		}

		// Task file (shortened if too long)
		taskFile := filepath.Base(entry.TaskFile)
		if len(taskFile) > maxTaskFileLen {
			taskFile = taskFile[:maxTaskFileLen-3] + "..."
		}

		// Executed at
		executedAt := entry.ExecutedAt.Format("2006-01-02 15:04:05")

		// Duration
		duration := formatDuration(entry.Duration)

		// Branch name (shortened if too long)
		branch := entry.BranchName
		if branch == "" {
			branch = "-"
		} else {
			// Remove "feature/" prefix for display
			branch = strings.TrimPrefix(branch, "feature/")
			if len(branch) > 20 {
				branch = branch[:17] + "..."
			}
		}

		fmt.Printf(entryFormat, status, taskFile, executedAt, duration, entry.TaskCount, entry.MaxRetries, branch)

		// Show error message if failed
		if !entry.Success && entry.ErrorMessage != "" {
			errorMsg := entry.ErrorMessage
			if len(errorMsg) > 100 {
				errorMsg = errorMsg[:97] + "..."
			}
			fmt.Printf("       %s %s\n", yellow("Error:"), errorMsg)
		}
	}

	fmt.Println()

	// Summary statistics
	successCount := 0
	failedCount := 0
	totalDuration := time.Duration(0)

	for _, entry := range entries {
		if entry.Success {
			successCount++
		} else {
			failedCount++
		}
		totalDuration += entry.Duration
	}

	fmt.Printf("📊 Summary:\n")
	fmt.Printf("   Total: %d | ", len(entries))
	fmt.Printf("%s: %d | ", green("Success"), successCount)
	fmt.Printf("%s: %d | ", red("Failed"), failedCount)
	fmt.Printf("%s: %s\n", blue("Total Duration"), formatDuration(totalDuration))
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		minutes := int(d.Minutes())
		seconds := int(d.Seconds()) % 60
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	return fmt.Sprintf("%dh %dm", hours, minutes)
}
