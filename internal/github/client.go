package github

import (
	"context"
	"fmt"
	"log"

	"github.com/google/go-github/v57/github"
	"github.com/isiidaisuke0926/auto-issue-finder/pkg/models"
	"golang.org/x/oauth2"
)

// Client はGitHub APIクライアントをラップする
type Client struct {
	client  *github.Client
	ctx     context.Context
	verbose bool
}

// NewClient は新しいGitHubクライアントを作成する
func NewClient(token string, verbose bool) *Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	return &Client{
		client:  github.NewClient(tc),
		ctx:     ctx,
		verbose: verbose,
	}
}

// FetchIssues はページネーションを使用してリポジトリからIssueを取得する
func (c *Client) FetchIssues(owner, repo string, state string, labels []string, limit int) ([]models.IssueSummary, error) {
	var allIssues []models.IssueSummary
	opts := &github.IssueListByRepoOptions{
		State:  state,
		Labels: labels,
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	page := 1
	for {
		if c.verbose {
			log.Printf("Fetching page %d...", page)
		}

		opts.Page = page
		issues, resp, err := c.client.Issues.ListByRepo(c.ctx, owner, repo, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch issues: %w", err)
		}

		for _, issue := range issues {
			// プルリクエストはスキップ
			if issue.PullRequestLinks != nil {
				continue
			}

			summary := c.convertToSummary(issue)
			allIssues = append(allIssues, summary)

			if limit > 0 && len(allIssues) >= limit {
				return allIssues[:limit], nil
			}
		}

		if resp.NextPage == 0 {
			break
		}
		page = resp.NextPage

		// レート制限の考慮
		if c.verbose {
			rate, _, err := c.client.RateLimit.Get(c.ctx)
			if err == nil {
				log.Printf("Rate limit: %d/%d remaining", rate.Core.Remaining, rate.Core.Limit)
			}
		}
	}

	return allIssues, nil
}

// GetRateLimit は現在のレート制限状態を返す
func (c *Client) GetRateLimit() (*github.RateLimits, error) {
	rate, _, err := c.client.RateLimit.Get(c.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get rate limit: %w", err)
	}
	return rate, nil
}

// convertToSummary はGitHub IssueをモデルIssueに変換する
func (c *Client) convertToSummary(issue *github.Issue) models.IssueSummary {
	summary := models.IssueSummary{
		Number:    issue.GetNumber(),
		Title:     issue.GetTitle(),
		State:     issue.GetState(),
		Comments:  issue.GetComments(),
		CreatedAt: issue.GetCreatedAt().Time,
		UpdatedAt: issue.GetUpdatedAt().Time,
		HTMLURL:   issue.GetHTMLURL(),
	}

	if issue.ClosedAt != nil {
		closedAt := issue.GetClosedAt().Time
		summary.ClosedAt = &closedAt
	}

	for _, label := range issue.Labels {
		summary.Labels = append(summary.Labels, label.GetName())
	}

	return summary
}

// ValidateToken はトークンが有効かどうかを確認する
func (c *Client) ValidateToken() error {
	_, _, err := c.client.Users.Get(c.ctx, "")
	if err != nil {
		return fmt.Errorf("invalid token: %w", err)
	}
	return nil
}
