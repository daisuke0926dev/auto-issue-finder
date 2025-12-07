package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/isiidaisuke0926/auto-issue-finder/internal/analyzer"
	githubclient "github.com/isiidaisuke0926/auto-issue-finder/internal/github"
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

func main() {
	// Load .env file if exists
	_ = godotenv.Load()

	var rootCmd = &cobra.Command{
		Use:   "auto-issue-finder",
		Short: "Analyze GitHub issues and generate insights",
		Long: `auto-issue-finder is a CLI tool that analyzes GitHub repository issues,
detects patterns, and provides actionable recommendations.`,
	}

	var analyzeCmd = &cobra.Command{
		Use:   "analyze [owner/repo]",
		Short: "Analyze issues in a GitHub repository",
		Long: `Fetch and analyze issues from a GitHub repository.
Generates statistics, detects patterns, and provides recommendations.

Examples:
  auto-issue-finder analyze microsoft/vscode
  auto-issue-finder analyze golang/go --state=open --limit=100
  auto-issue-finder analyze owner/repo --format=json --output=report.json`,
		Args: cobra.ExactArgs(1),
		Run:  runAnalyze,
	}

	// Define flags
	analyzeCmd.Flags().StringVar(&token, "token", os.Getenv("GITHUB_TOKEN"), "GitHub personal access token (or set GITHUB_TOKEN env var)")
	analyzeCmd.Flags().StringVar(&state, "state", "all", "Filter by state: open, closed, or all")
	analyzeCmd.Flags().StringSliceVar(&labels, "labels", []string{}, "Filter by labels (comma-separated)")
	analyzeCmd.Flags().StringVar(&format, "format", "markdown", "Output format: markdown, json, or console")
	analyzeCmd.Flags().StringVar(&output, "output", "", "Output file path (default: stdout)")
	analyzeCmd.Flags().IntVar(&limit, "limit", 0, "Maximum number of issues to fetch (0 = all)")
	analyzeCmd.Flags().BoolVar(&verbose, "verbose", false, "Enable verbose logging")

	rootCmd.AddCommand(analyzeCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runAnalyze(cmd *cobra.Command, args []string) {
	// Parse repository
	repoParts := strings.Split(args[0], "/")
	if len(repoParts) != 2 {
		log.Fatal("Repository must be in format: owner/repo")
	}
	owner, repo := repoParts[0], repoParts[1]

	// Validate token
	if token == "" {
		log.Fatal("GitHub token is required. Set GITHUB_TOKEN environment variable or use --token flag")
	}

	// Create GitHub client
	if verbose {
		log.Println("Creating GitHub client...")
	}
	client := githubclient.NewClient(token, verbose)

	// Validate token
	if err := client.ValidateToken(); err != nil {
		log.Fatalf("Token validation failed: %v", err)
	}

	// Fetch issues
	if format == "console" || output == "" {
		fmt.Printf("🔍 Analyzing %s/%s...\n", owner, repo)
	}

	issues, err := client.FetchIssues(owner, repo, state, labels, limit)
	if err != nil {
		log.Fatalf("Failed to fetch issues: %v", err)
	}

	if len(issues) == 0 {
		log.Println("No issues found matching the criteria")
		return
	}

	if verbose {
		log.Printf("Fetched %d issues", len(issues))
	}

	// Analyze issues
	if verbose {
		log.Println("Analyzing issues...")
	}
	a := analyzer.NewAnalyzer(issues)
	result := a.Analyze()
	recommendations := a.GenerateRecommendations(result)

	// Generate report
	r := reporter.NewReporter(result, args[0])
	var reportContent string

	switch format {
	case "json":
		jsonReport, err := r.GenerateJSON()
		if err != nil {
			log.Fatalf("Failed to generate JSON report: %v", err)
		}
		reportContent = jsonReport
	case "console":
		reportContent = r.GenerateConsoleOutput(recommendations)
	case "markdown":
		reportContent = r.GenerateMarkdown(recommendations)
	default:
		log.Fatalf("Unknown format: %s (use markdown, json, or console)", format)
	}

	// Output report
	if output != "" {
		err := os.WriteFile(output, []byte(reportContent), 0644)
		if err != nil {
			log.Fatalf("Failed to write output file: %v", err)
		}
		fmt.Printf("📝 Report saved to: %s\n", output)
	} else {
		fmt.Print(reportContent)
	}

	// Show summary if saving to file
	if output != "" && format != "console" {
		fmt.Printf("\n✓ Analyzed %d issues\n", len(issues))
		fmt.Printf("✓ Found %d recommendations\n", len(recommendations))
	}
}
