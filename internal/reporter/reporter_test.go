package reporter

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/isiidaisuke0926/auto-issue-finder/pkg/models"
	"github.com/stretchr/testify/assert"
)

func createTestResult() *models.AnalysisResult {
	now := time.Now()
	_ = now.AddDate(0, 0, -1) // closedTime not needed in this test

	return &models.AnalysisResult{
		Repository: "test/repo",
		Stats: models.IssueStats{
			TotalIssues:       10,
			OpenIssues:        7,
			ClosedIssues:      3,
			AvgResolutionDays: 5.5,
			LabelDistribution: map[string]int{
				"bug":         4,
				"enhancement": 3,
				"question":    2,
				"documentation": 1,
			},
			MonthlyDistribution: map[string]int{
				"2024-11": 5,
				"2024-12": 5,
			},
		},
		Patterns: models.IssuePattern{
			Keywords: map[string]int{
				"authentication": 3,
				"feature":        2,
				"error":          2,
			},
			LongStandingIssues: []models.IssueSummary{
				{
					Number:    1,
					Title:     "Old issue",
					Comments:  5,
					CreatedAt: now.AddDate(0, 0, -60),
					HTMLURL:   "https://github.com/test/repo/issues/1",
				},
			},
			HotTopics: []models.IssueSummary{
				{
					Number:   2,
					Title:    "Popular issue",
					Comments: 25,
					HTMLURL:  "https://github.com/test/repo/issues/2",
				},
			},
			UnlabeledIssues: []models.IssueSummary{
				{
					Number: 3,
					Title:  "No labels",
				},
			},
		},
		Problems: models.IssueProblem{
			BugRatio: 0.4,
			PotentialDuplicates: []models.DuplicatePair{
				{
					Issue1: models.IssueSummary{
						Number: 4,
						Title:  "Duplicate 1",
					},
					Issue2: models.IssueSummary{
						Number: 5,
						Title:  "Duplicate 2",
					},
					Similarity: 0.75,
				},
			},
			StaleIssues: []models.IssueSummary{
				{
					Number:    6,
					Title:     "Stale issue",
					UpdatedAt: now.AddDate(0, 0, -20),
				},
			},
		},
		GeneratedAt: now,
	}
}

func TestNewReporter(t *testing.T) {
	result := createTestResult()
	reporter := NewReporter(result, "test/repo")

	assert.NotNil(t, reporter)
	assert.Equal(t, "test/repo", reporter.repo)
	assert.Equal(t, "test/repo", reporter.result.Repository)
}

func TestGenerateMarkdown(t *testing.T) {
	result := createTestResult()
	reporter := NewReporter(result, "test/repo")

	recommendations := []models.Recommendation{
		{
			Priority:    "High",
			Category:    "Organization",
			Description: "Review unlabeled issues",
			Count:       1,
		},
		{
			Priority:    "Medium",
			Category:    "Maintenance",
			Description: "Update stale issues",
			Count:       1,
		},
	}

	markdown := reporter.GenerateMarkdown(recommendations)

	assert.NotEmpty(t, markdown)

	// Check for key sections
	assert.Contains(t, markdown, "# GitHub Issue Analysis Report")
	assert.Contains(t, markdown, "test/repo")
	assert.Contains(t, markdown, "📊 Issue Statistics")
	assert.Contains(t, markdown, "Total Issues")
	assert.Contains(t, markdown, "10")

	// Check for label distribution
	assert.Contains(t, markdown, "📋 Label Distribution")
	assert.Contains(t, markdown, "bug")

	// Check for monthly trend
	assert.Contains(t, markdown, "📈 Monthly Issue Creation Trend")

	// Check for attention items
	assert.Contains(t, markdown, "⚠️  Issues Needing Attention")

	// Check for recommendations
	assert.Contains(t, markdown, "💡 Recommendations")
	assert.Contains(t, markdown, "Review unlabeled issues")
}

func TestGenerateJSON(t *testing.T) {
	result := createTestResult()
	reporter := NewReporter(result, "test/repo")

	jsonStr, err := reporter.GenerateJSON()

	assert.NoError(t, err)
	assert.NotEmpty(t, jsonStr)

	// Validate JSON structure
	var parsed models.AnalysisResult
	err = json.Unmarshal([]byte(jsonStr), &parsed)
	assert.NoError(t, err)

	assert.Equal(t, "test/repo", parsed.Repository)
	assert.Equal(t, 10, parsed.Stats.TotalIssues)
	assert.Equal(t, 7, parsed.Stats.OpenIssues)
	assert.Equal(t, 4, parsed.Stats.LabelDistribution["bug"])
}

func TestGenerateConsoleOutput(t *testing.T) {
	result := createTestResult()
	reporter := NewReporter(result, "test/repo")

	recommendations := []models.Recommendation{
		{
			Priority:    "High",
			Category:    "Organization",
			Description: "Review unlabeled issues",
			Count:       1,
		},
	}

	output := reporter.GenerateConsoleOutput(recommendations)

	assert.NotEmpty(t, output)

	// Check for emojis and sections
	assert.Contains(t, output, "🔍 Analyzing")
	assert.Contains(t, output, "test/repo")
	assert.Contains(t, output, "📊 Issue Statistics")
	assert.Contains(t, output, "Total Issues:")
	assert.Contains(t, output, "10")

	// Check for separator lines
	assert.Contains(t, output, "─────────────────────────────────")

	// Check for recommendations
	assert.Contains(t, output, "💡 Recommendations")
	assert.Contains(t, output, "Review unlabeled issues")
}

func TestMarkdownWithNoIssues(t *testing.T) {
	result := &models.AnalysisResult{
		Repository: "test/repo",
		Stats: models.IssueStats{
			TotalIssues:         0,
			OpenIssues:          0,
			ClosedIssues:        0,
			LabelDistribution:   map[string]int{},
			MonthlyDistribution: map[string]int{},
		},
		Patterns: models.IssuePattern{
			Keywords:           map[string]int{},
			LongStandingIssues: []models.IssueSummary{},
			HotTopics:          []models.IssueSummary{},
			UnlabeledIssues:    []models.IssueSummary{},
		},
		Problems: models.IssueProblem{
			BugRatio:            0.0,
			PotentialDuplicates: []models.DuplicatePair{},
			StaleIssues:         []models.IssueSummary{},
		},
		GeneratedAt: time.Now(),
	}

	reporter := NewReporter(result, "test/repo")
	markdown := reporter.GenerateMarkdown([]models.Recommendation{})

	assert.NotEmpty(t, markdown)
	assert.Contains(t, markdown, "No significant issues detected")
}

func TestConsoleOutputWithNoRecommendations(t *testing.T) {
	result := createTestResult()
	reporter := NewReporter(result, "test/repo")

	output := reporter.GenerateConsoleOutput([]models.Recommendation{})

	assert.NotEmpty(t, output)
	assert.Contains(t, output, "test/repo")
	assert.Contains(t, output, "📊 Issue Statistics")
}

func TestLabelDistributionSorting(t *testing.T) {
	result := &models.AnalysisResult{
		Stats: models.IssueStats{
			TotalIssues: 100,
			LabelDistribution: map[string]int{
				"z-label": 1,
				"a-label": 50,
				"m-label": 25,
				"b-label": 10,
			},
		},
		Patterns:    models.IssuePattern{Keywords: map[string]int{}},
		Problems:    models.IssueProblem{},
		GeneratedAt: time.Now(),
	}

	reporter := NewReporter(result, "test/repo")
	markdown := reporter.GenerateMarkdown([]models.Recommendation{})

	// Labels should be sorted by count (descending)
	aLabelPos := strings.Index(markdown, "a-label")
	mLabelPos := strings.Index(markdown, "m-label")
	bLabelPos := strings.Index(markdown, "b-label")

	assert.Less(t, aLabelPos, mLabelPos, "a-label (50) should appear before m-label (25)")
	assert.Less(t, mLabelPos, bLabelPos, "m-label (25) should appear before b-label (10)")
}

func TestMonthlyTrendChart(t *testing.T) {
	result := createTestResult()
	reporter := NewReporter(result, "test/repo")

	markdown := reporter.GenerateMarkdown([]models.Recommendation{})

	// Should contain ASCII chart
	assert.Contains(t, markdown, "```")
	assert.Contains(t, markdown, "2024-11")
	assert.Contains(t, markdown, "2024-12")
}
