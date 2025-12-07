# Auto Issue Finder

A powerful CLI tool that automatically analyzes GitHub repository issues, detects patterns, and provides actionable recommendations.

## Features

- Fetch issues from any GitHub repository with pagination support
- Comprehensive issue analysis:
  - Basic statistics (open/closed, average resolution time)
  - Label distribution and trends
  - Monthly issue creation patterns
  - Keyword extraction from issue titles
- Pattern detection:
  - Long-standing issues (open >30 days)
  - Hot topics (issues with >20 comments)
  - Unlabeled issues
  - Stale issues (not updated in 14+ days)
- Problem identification:
  - Bug ratio analysis
  - Potential duplicate detection
  - Stale issue detection
- Multiple output formats:
  - Console (formatted for terminal)
  - Markdown (detailed reports)
  - JSON (for automation)
- Smart recommendations based on analysis

## Installation

### Prerequisites

- Go 1.21 or higher
- GitHub Personal Access Token

### Build from source

```bash
# Clone the repository
git clone https://github.com/isiidaisuke0926/auto-issue-finder.git
cd auto-issue-finder

# Install dependencies
go mod download

# Build the CLI tool
go build -o auto-issue-finder cmd/analyze/main.go

# (Optional) Install globally
go install cmd/analyze/main.go
```

## Quick Start

### 1. Set up GitHub Token

Create a `.env` file in the project root:

```bash
cp .env.example .env
# Edit .env and add your token
echo "GITHUB_TOKEN=your_github_token_here" > .env
```

Or set it as an environment variable:

```bash
export GITHUB_TOKEN=your_github_token_here
```

### 2. Run Analysis

```bash
# Analyze a repository (console output)
./auto-issue-finder analyze microsoft/vscode --format=console

# Generate markdown report
./auto-issue-finder analyze golang/go --format=markdown --output=report.md

# Limit to 100 issues
./auto-issue-finder analyze owner/repo --limit=100

# Filter by state and labels
./auto-issue-finder analyze owner/repo --state=open --labels=bug,enhancement

# JSON output for automation
./auto-issue-finder analyze owner/repo --format=json --output=analysis.json
```

## Usage

### Basic Command

```bash
auto-issue-finder analyze [owner/repo] [flags]
```

### Available Flags

| Flag | Description | Default | Example |
|------|-------------|---------|---------|
| `--token` | GitHub personal access token | `$GITHUB_TOKEN` | `--token=ghp_xxx` |
| `--state` | Filter by issue state | `all` | `--state=open` |
| `--labels` | Filter by labels (comma-separated) | `[]` | `--labels=bug,help-wanted` |
| `--format` | Output format | `markdown` | `--format=json` |
| `--output` | Output file path | stdout | `--output=report.md` |
| `--limit` | Max issues to fetch (0 = all) | `0` | `--limit=100` |
| `--verbose` | Enable verbose logging | `false` | `--verbose` |

### Examples

#### Example 1: Quick Console Analysis

```bash
./auto-issue-finder analyze microsoft/vscode --format=console --limit=50
```

Output:
```
🔍 Analyzing microsoft/vscode...
✓ Fetched 50 issues

📊 Issue Statistics
─────────────────────────────────
Total Issues:        50
Open:                42 (84%)
Closed:              8 (16%)
Avg Resolution Time: 12.3 days

📋 Label Distribution
─────────────────────────────────
bug                  23 (46%)
feature-request      15 (30%)
enhancement          8 (16%)

⚠️  Issues Needing Attention
─────────────────────────────────
• 12 issues without labels
• 18 issues open for >30 days
• 5 issues with >20 comments

💡 Recommendations
─────────────────────────────────
1. Consider triaging 12 unlabeled issues
2. Review 18 long-standing open issues
3. High activity issues may need prioritization (5)
```

#### Example 2: Generate Markdown Report

```bash
./auto-issue-finder analyze golang/go \
  --state=open \
  --format=markdown \
  --output=golang-issues.md
```

#### Example 3: JSON for Automation

```bash
./auto-issue-finder analyze owner/repo \
  --format=json \
  --output=analysis.json

# Use with jq for processing
cat analysis.json | jq '.Stats.TotalIssues'
```

## Output Formats

### Console Format

Optimized for terminal viewing with emojis and formatted sections.

### Markdown Format

Detailed report with:
- Statistics table
- Label distribution
- Monthly trend chart (ASCII)
- Long-standing issues list
- High-activity issues list
- Prioritized recommendations

See [examples/sample-report.md](examples/sample-report.md) for a sample.

### JSON Format

Complete analysis data in JSON format for programmatic processing.

## GitHub Token

### Creating a Token

1. Go to GitHub Settings > Developer settings > Personal access tokens > Tokens (classic)
2. Click "Generate new token"
3. Select scopes: `public_repo` (for public repos) or `repo` (for private repos)
4. Copy the token

### Setting the Token

Option 1: Environment variable
```bash
export GITHUB_TOKEN=ghp_xxxxxxxxxxxx
```

Option 2: `.env` file
```bash
echo "GITHUB_TOKEN=ghp_xxxxxxxxxxxx" > .env
```

Option 3: Command flag
```bash
./auto-issue-finder analyze owner/repo --token=ghp_xxxxxxxxxxxx
```

## Development

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test ./... -cover

# Run tests verbosely
go test ./... -v

# Check coverage for specific package
go test ./internal/analyzer -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Test Coverage

- `internal/analyzer`: 96.9%
- `internal/reporter`: 96.5%
- Overall: >70%

### Project Structure

```
auto-issue-finder/
├── cmd/
│   └── analyze/          # Main CLI command
│       └── main.go
├── internal/
│   ├── github/           # GitHub API client
│   │   ├── client.go
│   │   └── client_test.go
│   ├── analyzer/         # Issue analysis logic
│   │   ├── analyzer.go
│   │   └── analyzer_test.go
│   └── reporter/         # Report generation
│       ├── reporter.go
│       └── reporter_test.go
├── pkg/
│   └── models/           # Data models
│       └── models.go
├── examples/             # Sample outputs
├── go.mod
├── go.sum
├── README.md
└── CONTRIBUTING.md
```

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details.

### Quick Start for Contributors

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests (`go test ./...`)
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

## Troubleshooting

### "invalid token" Error

Make sure your GitHub token:
- Has the correct scopes (`repo` or `public_repo`)
- Is not expired
- Is correctly set in `.env` or environment variable

### Rate Limiting

GitHub API has rate limits:
- Authenticated: 5,000 requests/hour
- Unauthenticated: 60 requests/hour

Use `--verbose` to see rate limit status:

```bash
./auto-issue-finder analyze owner/repo --verbose
```

### No Issues Found

Check:
- Repository exists and is public (or you have access)
- State filter matches issues (`--state=all` to see all)
- Label filters are correct

## Roadmap

- [ ] Cache support for faster re-analysis
- [ ] HTML dashboard output
- [ ] Multi-repository batch analysis
- [ ] GitHub Actions integration
- [ ] Issue timeline analysis
- [ ] Contributor statistics
- [ ] Custom analysis rules

## License

MIT License - see [LICENSE](LICENSE) for details

## Credits

Built with:
- [go-github](https://github.com/google/go-github) - GitHub API client
- [cobra](https://github.com/spf13/cobra) - CLI framework
- [godotenv](https://github.com/joho/godotenv) - Environment variable loader
- [testify](https://github.com/stretchr/testify) - Testing toolkit

## Support

- Report bugs: [GitHub Issues](https://github.com/isiidaisuke0926/auto-issue-finder/issues)
- Questions: [GitHub Discussions](https://github.com/isiidaisuke0926/auto-issue-finder/discussions)

---

**Happy analyzing!**
