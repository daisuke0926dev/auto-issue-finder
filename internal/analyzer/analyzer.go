package analyzer

import (
	"sort"
	"strings"
	"time"

	"github.com/isiidaisuke0926/auto-issue-finder/pkg/models"
)

// Analyzer analyzes issues and generates insights
type Analyzer struct {
	issues []models.IssueSummary
}

// NewAnalyzer creates a new analyzer
func NewAnalyzer(issues []models.IssueSummary) *Analyzer {
	return &Analyzer{
		issues: issues,
	}
}

// Analyze performs comprehensive analysis
func (a *Analyzer) Analyze() *models.AnalysisResult {
	return &models.AnalysisResult{
		Stats:       a.calculateStats(),
		Patterns:    a.detectPatterns(),
		Problems:    a.detectProblems(),
		GeneratedAt: time.Now(),
	}
}

// calculateStats computes basic statistics
func (a *Analyzer) calculateStats() models.IssueStats {
	stats := models.IssueStats{
		LabelDistribution:   make(map[string]int),
		MonthlyDistribution: make(map[string]int),
	}

	var totalResolutionDays float64
	var closedCount int

	for _, issue := range a.issues {
		stats.TotalIssues++

		if issue.State == "open" {
			stats.OpenIssues++
		} else {
			stats.ClosedIssues++
			if issue.ClosedAt != nil {
				days := issue.ClosedAt.Sub(issue.CreatedAt).Hours() / 24
				totalResolutionDays += days
				closedCount++
			}
		}

		for _, label := range issue.Labels {
			stats.LabelDistribution[label]++
		}

		month := issue.CreatedAt.Format("2006-01")
		stats.MonthlyDistribution[month]++
	}

	if closedCount > 0 {
		stats.AvgResolutionDays = totalResolutionDays / float64(closedCount)
	}

	return stats
}

// detectPatterns identifies patterns in issues
func (a *Analyzer) detectPatterns() models.IssuePattern {
	patterns := models.IssuePattern{
		Keywords: make(map[string]int),
	}

	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true,
		"but": true, "in": true, "on": true, "at": true, "to": true,
		"for": true, "of": true, "with": true, "by": true, "from": true,
		"is": true, "are": true, "was": true, "were": true, "be": true,
		"been": true, "being": true, "have": true, "has": true, "had": true,
		"do": true, "does": true, "did": true, "will": true, "would": true,
		"could": true, "should": true, "may": true, "might": true, "can": true,
	}

	for _, issue := range a.issues {
		words := strings.Fields(strings.ToLower(issue.Title))
		for _, word := range words {
			word = strings.Trim(word, ".,!?;:()[]{}\"'")
			if len(word) > 3 && !stopWords[word] {
				patterns.Keywords[word]++
			}
		}

		if issue.State == "open" {
			daysSinceCreation := time.Since(issue.CreatedAt).Hours() / 24
			if daysSinceCreation > 30 { // 30 days threshold for long-standing issues
				patterns.LongStandingIssues = append(patterns.LongStandingIssues, issue)
			}
		}

		if issue.Comments > 20 {
			patterns.HotTopics = append(patterns.HotTopics, issue)
		}

		if len(issue.Labels) == 0 {
			patterns.UnlabeledIssues = append(patterns.UnlabeledIssues, issue)
		}
	}

	return patterns
}

// detectProblems identifies potential problems
func (a *Analyzer) detectProblems() models.IssueProblem {
	problems := models.IssueProblem{}

	var bugCount int
	for _, issue := range a.issues {
		for _, label := range issue.Labels {
			if strings.Contains(strings.ToLower(label), "bug") {
				bugCount++
				break
			}
		}

		daysSinceUpdate := time.Since(issue.UpdatedAt).Hours() / 24
		if issue.State == "open" && daysSinceUpdate > 14 { // 14 days threshold for stale issues
			problems.StaleIssues = append(problems.StaleIssues, issue)
		}
	}

	if len(a.issues) > 0 {
		problems.BugRatio = float64(bugCount) / float64(len(a.issues))
	}

	problems.PotentialDuplicates = a.findPotentialDuplicates()

	return problems
}

// findPotentialDuplicates identifies issues with similar titles
func (a *Analyzer) findPotentialDuplicates() []models.DuplicatePair {
	var duplicates []models.DuplicatePair

	for i := 0; i < len(a.issues); i++ {
		for j := i + 1; j < len(a.issues); j++ {
			similarity := a.calculateSimilarity(a.issues[i].Title, a.issues[j].Title)
			if similarity > 0.6 { // 60% similarity threshold
				duplicates = append(duplicates, models.DuplicatePair{
					Issue1:     a.issues[i],
					Issue2:     a.issues[j],
					Similarity: similarity,
				})
			}
		}
	}

	return duplicates
}

// calculateSimilarity calculates simple word overlap similarity
func (a *Analyzer) calculateSimilarity(title1, title2 string) float64 {
	words1 := strings.Fields(strings.ToLower(title1))
	words2 := strings.Fields(strings.ToLower(title2))

	if len(words1) == 0 || len(words2) == 0 {
		return 0
	}

	wordSet1 := make(map[string]bool)
	for _, word := range words1 {
		wordSet1[word] = true
	}

	overlap := 0
	for _, word := range words2 {
		if wordSet1[word] {
			overlap++
		}
	}

	maxLen := len(words1)
	if len(words2) > maxLen {
		maxLen = len(words2)
	}

	return float64(overlap) / float64(maxLen)
}

// KeywordCount represents a keyword and its count
type KeywordCount struct {
	Word  string
	Count int
}

// GetTopKeywords returns the most common keywords
func (a *Analyzer) GetTopKeywords(patterns models.IssuePattern, n int) []KeywordCount {
	var kvList []KeywordCount
	for word, count := range patterns.Keywords {
		kvList = append(kvList, KeywordCount{Word: word, Count: count})
	}

	sort.Slice(kvList, func(i, j int) bool {
		return kvList[i].Count > kvList[j].Count
	})

	if len(kvList) > n {
		kvList = kvList[:n]
	}

	return kvList
}

// GenerateRecommendations creates actionable recommendations based on analysis results
func (a *Analyzer) GenerateRecommendations(result *models.AnalysisResult) []models.Recommendation {
	var recommendations []models.Recommendation

	if len(result.Patterns.UnlabeledIssues) > 0 {
		recommendations = append(recommendations, models.Recommendation{
			Priority:    "High",
			Category:    "Organization",
			Description: "Consider triaging unlabeled issues",
			Count:       len(result.Patterns.UnlabeledIssues),
		})
	}

	if len(result.Patterns.LongStandingIssues) > 0 {
		recommendations = append(recommendations, models.Recommendation{
			Priority:    "Medium",
			Category:    "Maintenance",
			Description: "Review long-standing open issues",
			Count:       len(result.Patterns.LongStandingIssues),
		})
	}

	if len(result.Patterns.HotTopics) > 0 {
		recommendations = append(recommendations, models.Recommendation{
			Priority:    "High",
			Category:    "Engagement",
			Description: "High activity issues may need prioritization",
			Count:       len(result.Patterns.HotTopics),
		})
	}

	if len(result.Problems.StaleIssues) > 0 {
		recommendations = append(recommendations, models.Recommendation{
			Priority:    "Low",
			Category:    "Maintenance",
			Description: "Consider closing or updating stale issues",
			Count:       len(result.Problems.StaleIssues),
		})
	}

	if result.Problems.BugRatio > 0.5 {
		recommendations = append(recommendations, models.Recommendation{
			Priority:    "High",
			Category:    "Quality",
			Description: "High proportion of bugs detected",
			Count:       int(result.Problems.BugRatio * float64(result.Stats.TotalIssues)),
		})
	}

	if len(result.Problems.PotentialDuplicates) > 0 {
		recommendations = append(recommendations, models.Recommendation{
			Priority:    "Medium",
			Category:    "Organization",
			Description: "Review potential duplicate issues",
			Count:       len(result.Problems.PotentialDuplicates),
		})
	}

	return recommendations
}
