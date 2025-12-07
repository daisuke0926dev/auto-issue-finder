# Usage Examples

This document provides practical examples of using the auto-issue-finder CLI tool.

## Basic Usage

### 1. Analyze a Repository (Console Output)

```bash
./auto-issue-finder analyze golang/go --limit=10 --format=console
```

Output:
```
🔍 Analyzing golang/go...
✓ Fetched 10 issues

📊 Issue Statistics
─────────────────────────────────
Total Issues:        10
Open:                8 (80%)
Closed:              2 (20%)
Avg Resolution Time: 5.2 days

📋 Label Distribution
─────────────────────────────────
NeedsInvestigation   5 (50%)
Documentation        3 (30%)
Performance          2 (20%)

⚠️  Issues Needing Attention
─────────────────────────────────
• 2 issues without labels
• 3 issues open for >30 days
• 1 issues with >20 comments

💡 Recommendations
─────────────────────────────────
1. Consider triaging unlabeled issues (2)
2. Review long-standing open issues (3)
```

### 2. Generate Markdown Report

```bash
./auto-issue-finder analyze microsoft/vscode \
  --limit=50 \
  --format=markdown \
  --output=vscode-analysis.md
```

### 3. Filter by State

```bash
# Only open issues
./auto-issue-finder analyze owner/repo --state=open

# Only closed issues
./auto-issue-finder analyze owner/repo --state=closed

# All issues (default)
./auto-issue-finder analyze owner/repo --state=all
```

### 4. Filter by Labels

```bash
# Single label
./auto-issue-finder analyze owner/repo --labels=bug

# Multiple labels
./auto-issue-finder analyze owner/repo --labels=bug,enhancement,help-wanted
```

### 5. JSON Output for Automation

```bash
./auto-issue-finder analyze owner/repo \
  --format=json \
  --output=analysis.json

# Process with jq
cat analysis.json | jq '.Stats.TotalIssues'
cat analysis.json | jq '.Patterns.LongStandingIssues | length'
```

## Advanced Examples

### 6. Verbose Mode for Debugging

```bash
./auto-issue-finder analyze owner/repo --verbose
```

Output includes:
- Rate limit information
- Page fetching progress
- Detailed logging

### 7. Large Repository Analysis

```bash
# Fetch all issues (may take time)
./auto-issue-finder analyze kubernetes/kubernetes \
  --format=markdown \
  --output=k8s-full-analysis.md

# Limited fetch for quick analysis
./auto-issue-finder analyze kubernetes/kubernetes \
  --limit=200 \
  --format=console
```

### 8. Specific Use Cases

#### Finding Issues Needing Attention

```bash
# Open issues without labels
./auto-issue-finder analyze owner/repo \
  --state=open \
  --format=markdown \
  --output=unlabeled-report.md

# Review the "Issues Needing Attention" section
```

#### Bug Analysis

```bash
# Analyze only bug-labeled issues
./auto-issue-finder analyze owner/repo \
  --labels=bug \
  --format=console
```

#### Documentation Issues

```bash
./auto-issue-finder analyze owner/repo \
  --labels=documentation,docs \
  --state=open \
  --format=markdown \
  --output=docs-issues.md
```

## Automation Examples

### 9. Weekly Report Generation (Cron Job)

```bash
#!/bin/bash
# weekly-report.sh

DATE=$(date +%Y-%m-%d)
REPO="owner/repo"
OUTPUT="reports/report-$DATE.md"

./auto-issue-finder analyze $REPO \
  --format=markdown \
  --output=$OUTPUT

echo "Report generated: $OUTPUT"
```

Add to crontab:
```
0 9 * * 1 /path/to/weekly-report.sh
```

### 10. CI/CD Integration

```yaml
# .github/workflows/issue-analysis.yml
name: Weekly Issue Analysis

on:
  schedule:
    - cron: '0 9 * * 1'  # Every Monday at 9 AM

jobs:
  analyze:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Build tool
        run: go build -o auto-issue-finder cmd/analyze/main.go
      
      - name: Run analysis
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        run: |
          ./auto-issue-finder analyze ${{ github.repository }} \
            --format=markdown \
            --output=issue-analysis.md
      
      - name: Upload report
        uses: actions/upload-artifact@v3
        with:
          name: issue-report
          path: issue-analysis.md
```

### 11. Multiple Repository Analysis

```bash
#!/bin/bash
# analyze-multiple-repos.sh

REPOS=(
  "golang/go"
  "kubernetes/kubernetes"
  "microsoft/vscode"
)

for repo in "${REPOS[@]}"; do
  echo "Analyzing $repo..."
  SAFE_NAME=$(echo $repo | tr '/' '-')
  ./auto-issue-finder analyze $repo \
    --limit=100 \
    --format=markdown \
    --output="reports/${SAFE_NAME}.md"
done

echo "All analyses complete!"
```

## Output Processing

### 12. Extract Statistics from JSON

```bash
# Get total issues
./auto-issue-finder analyze owner/repo --format=json | jq '.Stats.TotalIssues'

# Get open issue count
./auto-issue-finder analyze owner/repo --format=json | jq '.Stats.OpenIssues'

# Get average resolution time
./auto-issue-finder analyze owner/repo --format=json | jq '.Stats.AvgResolutionDays'

# Get top labels
./auto-issue-finder analyze owner/repo --format=json | \
  jq '.Stats.LabelDistribution | to_entries | sort_by(.value) | reverse | .[0:5]'
```

### 13. Compare Repositories

```bash
#!/bin/bash
# compare-repos.sh

analyze() {
  ./auto-issue-finder analyze $1 --format=json --limit=100 | \
    jq "{repo: \"$1\", total: .Stats.TotalIssues, open: .Stats.OpenIssues}"
}

echo "["
analyze "golang/go"
echo ","
analyze "rust-lang/rust"
echo "]" | jq
```

## Tips and Best Practices

### Rate Limiting

- Use `--limit` to avoid hitting rate limits
- GitHub API: 5,000 requests/hour with authentication
- Use `--verbose` to monitor rate limit status

### Performance

- Start with smaller limits (`--limit=50`) for quick analysis
- Use filters (`--state`, `--labels`) to reduce data
- Cache results locally if analyzing repeatedly

### Token Management

```bash
# Set token once per session
export GITHUB_TOKEN=your_token_here

# Or use .env file
echo "GITHUB_TOKEN=your_token" > .env
```

## Troubleshooting

### No Issues Found

```bash
# Check if repository is accessible
./auto-issue-finder analyze owner/repo --verbose

# Try without filters
./auto-issue-finder analyze owner/repo --state=all --limit=5
```

### Token Issues

```bash
# Verify token is set
./auto-issue-finder analyze owner/repo --verbose
# Should show authentication error if token is invalid
```

## Sample Workflows

### New Project Setup

```bash
# 1. Get overview
./auto-issue-finder analyze owner/new-project --format=console

# 2. Generate detailed report
./auto-issue-finder analyze owner/new-project \
  --format=markdown \
  --output=initial-analysis.md

# 3. Focus on issues needing attention
./auto-issue-finder analyze owner/new-project \
  --state=open \
  --format=markdown \
  --output=action-items.md
```

### Maintenance Review

```bash
# 1. Check for stale issues
./auto-issue-finder analyze owner/repo --state=open --format=console

# 2. Review bug situation
./auto-issue-finder analyze owner/repo --labels=bug --format=markdown

# 3. Generate comprehensive report
./auto-issue-finder analyze owner/repo --output=monthly-review.md
```
