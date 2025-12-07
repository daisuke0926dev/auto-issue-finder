# Contributing to Auto Issue Finder

Thank you for considering contributing to Auto Issue Finder! This document provides guidelines and instructions for contributing.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Running Tests](#running-tests)
- [Submitting Changes](#submitting-changes)
- [Coding Guidelines](#coding-guidelines)
- [Project Structure](#project-structure)

## Code of Conduct

This project adheres to a code of conduct that all contributors are expected to follow. Please be respectful and constructive in all interactions.

## Getting Started

### Prerequisites

- Go 1.21 or higher
- Git
- GitHub account
- GitHub Personal Access Token (for testing)

### Fork and Clone

1. Fork the repository on GitHub
2. Clone your fork:

```bash
git clone https://github.com/YOUR_USERNAME/auto-issue-finder.git
cd auto-issue-finder
```

3. Add the upstream repository:

```bash
git remote add upstream https://github.com/isiidaisuke0926/auto-issue-finder.git
```

## Development Setup

### 1. Install Dependencies

```bash
go mod download
```

### 2. Set Up Environment

Create a `.env` file for testing:

```bash
cp .env.example .env
# Edit .env and add your GitHub token
echo "GITHUB_TOKEN=your_test_token" > .env
```

### 3. Verify Installation

```bash
# Build the project
go build -o auto-issue-finder cmd/analyze/main.go

# Run tests
go test ./...

# Test the CLI
./auto-issue-finder analyze golang/go --limit=5 --format=console
```

## Running Tests

### Unit Tests

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test ./... -v

# Run tests with coverage
go test ./... -cover

# Generate coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Test a Specific Package

```bash
# Test analyzer package
go test ./internal/analyzer -v

# Test with coverage
go test ./internal/analyzer -coverprofile=coverage.out
go tool cover -func=coverage.out
```

## Submitting Changes

### 1. Create a Branch

```bash
git checkout -b feature/your-feature-name
```

Branch naming conventions:
- `feature/` - New features
- `fix/` - Bug fixes
- `docs/` - Documentation changes
- `refactor/` - Code refactoring
- `test/` - Test additions or modifications

### 2. Make Your Changes

- Write clear, concise code
- Follow the coding guidelines (see below)
- Add tests for new features
- Update documentation as needed

### 3. Test Your Changes

```bash
# Run all tests
go test ./...

# Check formatting
go fmt ./...

# Run linter (if golangci-lint is installed)
golangci-lint run
```

### 4. Commit Your Changes

Write clear commit messages following conventional commits format.

### 5. Push and Create Pull Request

```bash
git push origin feature/your-feature-name
```

## Coding Guidelines

### Go Code Style

- Follow [Effective Go](https://golang.org/doc/effective_go)
- Use `gofmt` for formatting
- Use meaningful variable names
- Add comments for exported functions and types
- Keep functions focused and small

### Testing Guidelines

- Write tests for all new features
- Aim for >70% code coverage
- Use table-driven tests where appropriate
- Use meaningful test names

## Project Structure

```
auto-issue-finder/
├── cmd/analyze/          # Main CLI application
├── internal/github/      # GitHub API integration
├── internal/analyzer/    # Issue analysis logic
├── internal/reporter/    # Report generation
└── pkg/models/          # Shared data structures
```

## Getting Help

- **Questions**: Open a GitHub Discussion
- **Bugs**: Open a GitHub Issue

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

Thank you for contributing to Auto Issue Finder!
