package reporter

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/isiidaisuke0926/auto-issue-finder/pkg/models"
)

// Reporter generates reports in various formats
type Reporter struct {
	result *models.AnalysisResult
	repo   string
}

// NewReporter creates a new reporter
func NewReporter(result *models.AnalysisResult, repo string) *Reporter {
	result.Repository = repo
	return &Reporter{
		result: result,
		repo:   repo,
	}
}

// GenerateMarkdown creates a Markdown report
func (r *Reporter) GenerateMarkdown(recommendations []models.Recommendation) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# GitHub Issue Analysis Report\n\n"))
	sb.WriteString(fmt.Sprintf("**Repository:** %s  \n", r.repo))
	sb.WriteString(fmt.Sprintf("**Generated:** %s  \n\n", r.result.GeneratedAt.Format("2006-01-02 15:04:05")))

	sb.WriteString("## 📊 Issue Statistics\n\n")
	sb.WriteString("| Metric | Value |\n")
	sb.WriteString("|--------|-------|\n")
	sb.WriteString(fmt.Sprintf("| Total Issues | %d |\n", r.result.Stats.TotalIssues))
	sb.WriteString(fmt.Sprintf("| Open | %d (%.1f%%) |\n",
		r.result.Stats.OpenIssues,
		float64(r.result.Stats.OpenIssues)/float64(r.result.Stats.TotalIssues)*100))
	sb.WriteString(fmt.Sprintf("| Closed | %d (%.1f%%) |\n",
		r.result.Stats.ClosedIssues,
		float64(r.result.Stats.ClosedIssues)/float64(r.result.Stats.TotalIssues)*100))
	sb.WriteString(fmt.Sprintf("| Avg Resolution Time | %.1f days |\n\n", r.result.Stats.AvgResolutionDays))

	if len(r.result.Stats.LabelDistribution) > 0 {
		sb.WriteString("## 📋 Label Distribution\n\n")
		sb.WriteString("| Label | Count | Percentage |\n")
		sb.WriteString("|-------|-------|------------|\n")

		type labelCount struct {
			Label string
			Count int
		}
		var labels []labelCount
		for label, count := range r.result.Stats.LabelDistribution {
			labels = append(labels, labelCount{Label: label, Count: count})
		}
		sort.Slice(labels, func(i, j int) bool {
			return labels[i].Count > labels[j].Count
		})

		for i, lc := range labels {
			if i >= 10 { // Top 10 only
				break
			}
			pct := float64(lc.Count) / float64(r.result.Stats.TotalIssues) * 100
			sb.WriteString(fmt.Sprintf("| %s | %d | %.1f%% |\n", lc.Label, lc.Count, pct))
		}
		sb.WriteString("\n")
	}

	if len(r.result.Stats.MonthlyDistribution) > 0 {
		sb.WriteString("## 📈 Monthly Issue Creation Trend\n\n")

		type monthCount struct {
			Month string
			Count int
		}
		var months []monthCount
		for month, count := range r.result.Stats.MonthlyDistribution {
			months = append(months, monthCount{Month: month, Count: count})
		}
		sort.Slice(months, func(i, j int) bool {
			return months[i].Month < months[j].Month
		})

		sb.WriteString("```\n")
		for _, mc := range months {
			bar := strings.Repeat("█", mc.Count/2)
			if mc.Count > 0 && len(bar) == 0 {
				bar = "▏"
			}
			sb.WriteString(fmt.Sprintf("%s: %s %d\n", mc.Month, bar, mc.Count))
		}
		sb.WriteString("```\n\n")
	}

	sb.WriteString("## ⚠️  Issues Needing Attention\n\n")

	hasIssues := false
	if len(r.result.Patterns.UnlabeledIssues) > 0 {
		sb.WriteString(fmt.Sprintf("- **%d unlabeled issues**\n", len(r.result.Patterns.UnlabeledIssues)))
		hasIssues = true
	}
	if len(r.result.Patterns.LongStandingIssues) > 0 {
		sb.WriteString(fmt.Sprintf("- **%d issues open for >30 days**\n", len(r.result.Patterns.LongStandingIssues)))
		hasIssues = true
	}
	if len(r.result.Patterns.HotTopics) > 0 {
		sb.WriteString(fmt.Sprintf("- **%d issues with >20 comments**\n", len(r.result.Patterns.HotTopics)))
		hasIssues = true
	}
	if len(r.result.Problems.StaleIssues) > 0 {
		sb.WriteString(fmt.Sprintf("- **%d stale issues** (not updated in 14+ days)\n", len(r.result.Problems.StaleIssues)))
		hasIssues = true
	}
	if len(r.result.Problems.PotentialDuplicates) > 0 {
		sb.WriteString(fmt.Sprintf("- **%d potential duplicate pairs**\n", len(r.result.Problems.PotentialDuplicates)))
		hasIssues = true
	}
	if r.result.Problems.BugRatio > 0.3 {
		bugCount := int(r.result.Problems.BugRatio * float64(r.result.Stats.TotalIssues))
		sb.WriteString(fmt.Sprintf("- **High bug ratio**: %.1f%% (%d bugs)\n", r.result.Problems.BugRatio*100, bugCount))
		hasIssues = true
	}

	if !hasIssues {
		sb.WriteString("*No significant issues detected.*\n")
	}
	sb.WriteString("\n")

	if len(r.result.Patterns.LongStandingIssues) > 0 && len(r.result.Patterns.LongStandingIssues) <= 10 {
		sb.WriteString("### Long-Standing Open Issues\n\n")
		for _, issue := range r.result.Patterns.LongStandingIssues {
			sb.WriteString(fmt.Sprintf("- [#%d](%s) %s (%d comments)\n",
				issue.Number, issue.HTMLURL, issue.Title, issue.Comments))
		}
		sb.WriteString("\n")
	}

	if len(r.result.Patterns.HotTopics) > 0 && len(r.result.Patterns.HotTopics) <= 10 {
		sb.WriteString("### High-Activity Issues\n\n")
		for _, issue := range r.result.Patterns.HotTopics {
			sb.WriteString(fmt.Sprintf("- [#%d](%s) %s (%d comments)\n",
				issue.Number, issue.HTMLURL, issue.Title, issue.Comments))
		}
		sb.WriteString("\n")
	}

	if len(recommendations) > 0 {
		sb.WriteString("## 💡 Recommendations\n\n")
		for i, rec := range recommendations {
			sb.WriteString(fmt.Sprintf("%d. **[%s - %s]** %s",
				i+1, rec.Priority, rec.Category, rec.Description))
			if rec.Count > 0 {
				sb.WriteString(fmt.Sprintf(" (%d items)", rec.Count))
			}
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	sb.WriteString("---\n")
	sb.WriteString("*Generated by auto-issue-finder*\n")

	return sb.String()
}

// GenerateJSON creates a JSON report
func (r *Reporter) GenerateJSON() (string, error) {
	data, err := json.MarshalIndent(r.result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return string(data), nil
}

// GenerateConsoleOutput creates formatted console output
func (r *Reporter) GenerateConsoleOutput(recommendations []models.Recommendation) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("\n🔍 Analyzing %s...\n", r.repo))
	sb.WriteString(fmt.Sprintf("✓ Fetched %d issues\n\n", r.result.Stats.TotalIssues))

	sb.WriteString("📊 Issue Statistics\n")
	sb.WriteString("─────────────────────────────────\n")
	sb.WriteString(fmt.Sprintf("Total Issues:        %d\n", r.result.Stats.TotalIssues))
	sb.WriteString(fmt.Sprintf("Open:                %d (%.0f%%)\n",
		r.result.Stats.OpenIssues,
		float64(r.result.Stats.OpenIssues)/float64(r.result.Stats.TotalIssues)*100))
	sb.WriteString(fmt.Sprintf("Closed:              %d (%.0f%%)\n",
		r.result.Stats.ClosedIssues,
		float64(r.result.Stats.ClosedIssues)/float64(r.result.Stats.TotalIssues)*100))
	sb.WriteString(fmt.Sprintf("Avg Resolution Time: %.1f days\n\n", r.result.Stats.AvgResolutionDays))

	if len(r.result.Stats.LabelDistribution) > 0 {
		sb.WriteString("📋 Label Distribution\n")
		sb.WriteString("─────────────────────────────────\n")

		type labelCount struct {
			Label string
			Count int
		}
		var labels []labelCount
		for label, count := range r.result.Stats.LabelDistribution {
			labels = append(labels, labelCount{Label: label, Count: count})
		}
		sort.Slice(labels, func(i, j int) bool {
			return labels[i].Count > labels[j].Count
		})

		for i, lc := range labels {
			if i >= 5 { // Top 5 only for console
				break
			}
			pct := float64(lc.Count) / float64(r.result.Stats.TotalIssues) * 100
			sb.WriteString(fmt.Sprintf("%-20s %3d (%.0f%%)\n", lc.Label, lc.Count, pct))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("⚠️  Issues Needing Attention\n")
	sb.WriteString("─────────────────────────────────\n")

	hasIssues := false
	if len(r.result.Patterns.UnlabeledIssues) > 0 {
		sb.WriteString(fmt.Sprintf("• %d issues without labels\n", len(r.result.Patterns.UnlabeledIssues)))
		hasIssues = true
	}
	if len(r.result.Patterns.LongStandingIssues) > 0 {
		sb.WriteString(fmt.Sprintf("• %d issues open for >30 days\n", len(r.result.Patterns.LongStandingIssues)))
		hasIssues = true
	}
	if len(r.result.Patterns.HotTopics) > 0 {
		sb.WriteString(fmt.Sprintf("• %d issues with >20 comments\n", len(r.result.Patterns.HotTopics)))
		hasIssues = true
	}

	if !hasIssues {
		sb.WriteString("• No significant issues detected\n")
	}
	sb.WriteString("\n")

	if len(recommendations) > 0 {
		sb.WriteString("💡 Recommendations\n")
		sb.WriteString("─────────────────────────────────\n")
		for i, rec := range recommendations {
			sb.WriteString(fmt.Sprintf("%d. %s", i+1, rec.Description))
			if rec.Count > 0 {
				sb.WriteString(fmt.Sprintf(" (%d)", rec.Count))
			}
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
