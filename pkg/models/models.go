package models

import "time"

// IssueStats はIssueに関する統計情報を表す
type IssueStats struct {
	TotalIssues         int
	OpenIssues          int
	ClosedIssues        int
	AvgResolutionDays   float64
	LabelDistribution   map[string]int
	MonthlyDistribution map[string]int
}

// IssuePattern はIssueから検出されたパターンを表す
type IssuePattern struct {
	Keywords           map[string]int
	LongStandingIssues []IssueSummary
	HotTopics          []IssueSummary
	UnlabeledIssues    []IssueSummary
}

// IssueProblem は検出された問題を表す
type IssueProblem struct {
	BugRatio            float64
	PotentialDuplicates []DuplicatePair
	StaleIssues         []IssueSummary
}

// IssueSummary は簡略化されたIssue表現
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

// DuplicatePair は重複の可能性があるIssueのペアを表す
type DuplicatePair struct {
	Issue1     IssueSummary
	Issue2     IssueSummary
	Similarity float64
}

// AnalysisResult は全ての分析結果を含む
type AnalysisResult struct {
	Repository string
	Stats      IssueStats
	Patterns   IssuePattern
	Problems   IssueProblem
	GeneratedAt time.Time
}

// Recommendation は改善提案を表す
type Recommendation struct {
	Priority    string
	Category    string
	Description string
	Count       int
}
