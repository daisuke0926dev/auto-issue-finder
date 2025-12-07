package analyzer

import (
	"testing"
	"time"

	"github.com/isiidaisuke0926/auto-issue-finder/pkg/models"
	"github.com/stretchr/testify/assert"
)

func createTestIssues() []models.IssueSummary {
	now := time.Now()
	thirtyDaysAgo := now.AddDate(0, 0, -30)
	sixtyDaysAgo := now.AddDate(0, 0, -60)
	twoWeeksAgo := now.AddDate(0, 0, -14)
	oneWeekAgo := now.AddDate(0, 0, -7)

	closedTime := now.AddDate(0, 0, -1)

	return []models.IssueSummary{
		{
			Number:    1,
			Title:     "Fix critical bug in authentication system",
			State:     "open",
			Labels:    []string{"bug", "priority-high"},
			Comments:  5,
			CreatedAt: thirtyDaysAgo,
			UpdatedAt: twoWeeksAgo,
			HTMLURL:   "https://github.com/test/repo/issues/1",
		},
		{
			Number:    2,
			Title:     "Add new feature for user dashboard",
			State:     "open",
			Labels:    []string{"enhancement", "feature-request"},
			Comments:  25,
			CreatedAt: sixtyDaysAgo,
			UpdatedAt: oneWeekAgo,
			HTMLURL:   "https://github.com/test/repo/issues/2",
		},
		{
			Number:    3,
			Title:     "Fix bug in payment processing",
			State:     "closed",
			Labels:    []string{"bug"},
			Comments:  3,
			CreatedAt: thirtyDaysAgo,
			UpdatedAt: oneWeekAgo,
			ClosedAt:  &closedTime,
			HTMLURL:   "https://github.com/test/repo/issues/3",
		},
		{
			Number:    4,
			Title:     "Update documentation for API endpoints",
			State:     "open",
			Labels:    []string{},
			Comments:  1,
			CreatedAt: now.AddDate(0, 0, -5),
			UpdatedAt: now.AddDate(0, 0, -5),
			HTMLURL:   "https://github.com/test/repo/issues/4",
		},
		{
			Number:    5,
			Title:     "Improve performance of search feature",
			State:     "open",
			Labels:    []string{"enhancement"},
			Comments:  10,
			CreatedAt: now.AddDate(0, 0, -45),
			UpdatedAt: now.AddDate(0, 0, -20),
			HTMLURL:   "https://github.com/test/repo/issues/5",
		},
	}
}

func TestNewAnalyzer(t *testing.T) {
	issues := createTestIssues()
	analyzer := NewAnalyzer(issues)

	assert.NotNil(t, analyzer)
	assert.Equal(t, len(issues), len(analyzer.issues))
}

func TestCalculateStats(t *testing.T) {
	issues := createTestIssues()
	analyzer := NewAnalyzer(issues)

	stats := analyzer.calculateStats()

	assert.Equal(t, 5, stats.TotalIssues)
	assert.Equal(t, 4, stats.OpenIssues)
	assert.Equal(t, 1, stats.ClosedIssues)
	assert.Greater(t, stats.AvgResolutionDays, 0.0)

	// Check label distribution
	assert.Equal(t, 2, stats.LabelDistribution["bug"])
	assert.Equal(t, 2, stats.LabelDistribution["enhancement"])
	assert.Equal(t, 1, stats.LabelDistribution["priority-high"])
	assert.Equal(t, 1, stats.LabelDistribution["feature-request"])

	// Check monthly distribution
	assert.NotEmpty(t, stats.MonthlyDistribution)
}

func TestDetectPatterns(t *testing.T) {
	issues := createTestIssues()
	analyzer := NewAnalyzer(issues)

	patterns := analyzer.detectPatterns()

	// Check keywords
	assert.NotEmpty(t, patterns.Keywords)
	assert.Greater(t, patterns.Keywords["feature"], 0)

	// Check long-standing issues (>30 days)
	assert.NotEmpty(t, patterns.LongStandingIssues)

	// Check hot topics (>20 comments)
	assert.Equal(t, 1, len(patterns.HotTopics))
	assert.Equal(t, 2, patterns.HotTopics[0].Number)

	// Check unlabeled issues
	assert.Equal(t, 1, len(patterns.UnlabeledIssues))
	assert.Equal(t, 4, patterns.UnlabeledIssues[0].Number)
}

func TestDetectProblems(t *testing.T) {
	issues := createTestIssues()
	analyzer := NewAnalyzer(issues)

	problems := analyzer.detectProblems()

	// Check bug ratio
	assert.Greater(t, problems.BugRatio, 0.0)
	assert.LessOrEqual(t, problems.BugRatio, 1.0)

	// Check stale issues (not updated in 14+ days)
	assert.NotEmpty(t, problems.StaleIssues)
}

func TestCalculateSimilarity(t *testing.T) {
	analyzer := NewAnalyzer([]models.IssueSummary{})

	tests := []struct {
		title1     string
		title2     string
		minSim     float64
		maxSim     float64
	}{
		{"Fix bug in authentication", "Fix bug in authentication", 1.0, 1.0},
		{"Fix bug in auth", "Fix bug in payment", 0.5, 0.8},
		{"Completely different title", "Nothing matches here", 0.0, 0.3},
		{"", "", 0.0, 0.0},
		{"Test", "Different", 0.0, 0.2},
	}

	for _, tt := range tests {
		sim := analyzer.calculateSimilarity(tt.title1, tt.title2)
		assert.GreaterOrEqual(t, sim, tt.minSim, "Similarity for '%s' and '%s'", tt.title1, tt.title2)
		assert.LessOrEqual(t, sim, tt.maxSim, "Similarity for '%s' and '%s'", tt.title1, tt.title2)
	}
}

func TestFindPotentialDuplicates(t *testing.T) {
	issues := []models.IssueSummary{
		{
			Number: 1,
			Title:  "Fix authentication bug in login system",
		},
		{
			Number: 2,
			Title:  "Fix authentication bug in login feature",
		},
		{
			Number: 3,
			Title:  "Add new dashboard feature",
		},
	}

	analyzer := NewAnalyzer(issues)
	duplicates := analyzer.findPotentialDuplicates()

	// Should find at least one potential duplicate (issues 1 and 2)
	assert.NotEmpty(t, duplicates)
	if len(duplicates) > 0 {
		assert.Greater(t, duplicates[0].Similarity, 0.6)
	}
}

func TestGetTopKeywords(t *testing.T) {
	issues := createTestIssues()
	analyzer := NewAnalyzer(issues)
	patterns := analyzer.detectPatterns()

	topKeywords := analyzer.GetTopKeywords(patterns, 5)

	assert.NotEmpty(t, topKeywords)
	assert.LessOrEqual(t, len(topKeywords), 5)

	// Keywords should be sorted by count (descending)
	for i := 0; i < len(topKeywords)-1; i++ {
		assert.GreaterOrEqual(t, topKeywords[i].Count, topKeywords[i+1].Count)
	}
}

func TestGenerateRecommendations(t *testing.T) {
	issues := createTestIssues()
	analyzer := NewAnalyzer(issues)
	result := analyzer.Analyze()

	recommendations := analyzer.GenerateRecommendations(result)

	assert.NotEmpty(t, recommendations)

	// Should have recommendation for unlabeled issues
	hasUnlabeledRec := false
	for _, rec := range recommendations {
		if rec.Category == "Organization" && rec.Count == 1 {
			hasUnlabeledRec = true
			break
		}
	}
	assert.True(t, hasUnlabeledRec)
}

func TestAnalyze(t *testing.T) {
	issues := createTestIssues()
	analyzer := NewAnalyzer(issues)

	result := analyzer.Analyze()

	assert.NotNil(t, result)
	assert.Equal(t, 5, result.Stats.TotalIssues)
	assert.NotEmpty(t, result.Patterns.Keywords)
	assert.NotZero(t, result.GeneratedAt)
}

func TestEmptyIssues(t *testing.T) {
	analyzer := NewAnalyzer([]models.IssueSummary{})

	result := analyzer.Analyze()

	assert.NotNil(t, result)
	assert.Equal(t, 0, result.Stats.TotalIssues)
	assert.Empty(t, result.Patterns.LongStandingIssues)
	assert.Empty(t, result.Patterns.HotTopics)
}
