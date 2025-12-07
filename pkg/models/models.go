package models

import "time"

// IssueStats represents statistical information about issues
type IssueStats struct {
	TotalIssues         int
	OpenIssues          int
	ClosedIssues        int
	AvgResolutionDays   float64
	LabelDistribution   map[string]int
	MonthlyDistribution map[string]int
}

// IssuePattern represents detected patterns in issues
type IssuePattern struct {
	Keywords           map[string]int
	LongStandingIssues []IssueSummary
	HotTopics          []IssueSummary
	UnlabeledIssues    []IssueSummary
}

// IssueProblem represents detected problems
type IssueProblem struct {
	BugRatio            float64
	PotentialDuplicates []DuplicatePair
	StaleIssues         []IssueSummary
}

// IssueSummary is a simplified issue representation
type IssueSummary struct {
	Number      int
	Title       string
	State       string
	Labels      []string
	Comments    int
	CreatedAt   time.Time
	UpdatedAt   time.Time
	ClosedAt    *time.Time
	HTMLURL     string
}

// DuplicatePair represents potentially duplicate issues
type DuplicatePair struct {
	Issue1     IssueSummary
	Issue2     IssueSummary
	Similarity float64
}

// AnalysisResult contains all analysis results
type AnalysisResult struct {
	Repository string
	Stats      IssueStats
	Patterns   IssuePattern
	Problems   IssueProblem
	GeneratedAt time.Time
}

// Recommendation represents an improvement suggestion
type Recommendation struct {
	Priority    string
	Category    string
	Description string
	Count       int
}
