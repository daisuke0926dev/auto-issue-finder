package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/isiidaisuke0926/auto-issue-finder/internal/analyzer"
	"github.com/isiidaisuke0926/auto-issue-finder/internal/github"
	"github.com/isiidaisuke0926/auto-issue-finder/internal/reporter"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

var (
	token   string
	state   string
	labels  []string
	format  string
	output  string
	limit   int
	verbose bool
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze [owner/repo]",
	Short: "Analyze issues in a GitHub repository",
	Long: `Analyze GitHub repository issues to identify patterns, statistics, and potential problems.

Examples:
  auto-issue-finder analyze microsoft/vscode
  auto-issue-finder analyze owner/repo --state=open --labels=bug,enhancement
  auto-issue-finder analyze owner/repo --format=json --output=report.json
  auto-issue-finder analyze owner/repo --limit=100 --verbose`,
	Args: cobra.ExactArgs(1),
	RunE: runAnalyze,
}

func init() {
	rootCmd.AddCommand(analyzeCmd)

	analyzeCmd.Flags().StringVar(&token, "token", "", "GitHub personal access token (or use GITHUB_TOKEN env var)")
	analyzeCmd.Flags().StringVar(&state, "state", "all", "Issue state: open, closed, or all")
	analyzeCmd.Flags().StringSliceVar(&labels, "labels", []string{}, "Filter by labels (comma-separated)")
	analyzeCmd.Flags().StringVar(&format, "format", "console", "Output format: console, markdown, or json")
	analyzeCmd.Flags().StringVar(&output, "output", "", "Output file path (default: stdout)")
	analyzeCmd.Flags().IntVar(&limit, "limit", 0, "Maximum number of issues to fetch (0 = no limit)")
	analyzeCmd.Flags().BoolVar(&verbose, "verbose", false, "Enable verbose logging")
}

func runAnalyze(cmd *cobra.Command, args []string) error {
	// Load .env file if it exists
	_ = godotenv.Load()

	// Get token
	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
	}
	if token == "" {
		return fmt.Errorf("GitHub token is required. Set GITHUB_TOKEN environment variable or use --token flag")
	}

	// Parse repository
	parts := strings.Split(args[0], "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid repository format. Use: owner/repo")
	}
	owner, repo := parts[0], parts[1]

	if verbose {
		log.SetOutput(os.Stderr)
		log.Printf("Analyzing repository: %s/%s", owner, repo)
		log.Printf("State: %s, Labels: %v, Limit: %d", state, labels, limit)
	} else {
		log.SetOutput(os.Stderr)
		log.SetFlags(0)
	}

	// Create GitHub client
	client := github.NewClient(token, verbose)

	// Validate token
	if err := client.ValidateToken(); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Fetch issues
	if !verbose && format == "console" {
		fmt.Fprintf(os.Stderr, "🔍 Fetching issues from %s/%s...\n", owner, repo)
	}

	issues, err := client.FetchIssues(owner, repo, state, labels, limit)
	if err != nil {
		return fmt.Errorf("failed to fetch issues: %w", err)
	}

	if len(issues) == 0 {
		return fmt.Errorf("no issues found matching the criteria")
	}

	if !verbose && format == "console" {
		fmt.Fprintf(os.Stderr, "✓ Fetched %d issues\n", len(issues))
	}

	// Analyze issues
	an := analyzer.NewAnalyzer(issues)
	result := an.Analyze()

	// Generate recommendations
	recommendations := an.GenerateRecommendations(result)

	// Generate report
	rep := reporter.NewReporter(result, args[0])
	var reportContent string

	switch format {
	case "console":
		reportContent = rep.GenerateConsoleOutput(recommendations)
	case "markdown":
		reportContent = rep.GenerateMarkdown(recommendations)
	case "json":
		jsonContent, err := rep.GenerateJSON()
		if err != nil {
			return fmt.Errorf("failed to generate JSON: %w", err)
		}
		reportContent = jsonContent
	default:
		return fmt.Errorf("invalid format: %s (use: console, markdown, or json)", format)
	}

	// Output report
	if output != "" {
		if err := os.WriteFile(output, []byte(reportContent), 0644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		if format == "console" {
			fmt.Fprintf(os.Stderr, "\n📝 Full report saved to: %s\n", output)
		} else {
			fmt.Fprintf(os.Stderr, "Report saved to: %s\n", output)
		}
	} else {
		fmt.Print(reportContent)
	}

	return nil
}
