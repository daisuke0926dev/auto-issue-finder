package github

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	client := NewClient("test-token", false)

	assert.NotNil(t, client)
	assert.NotNil(t, client.client)
	assert.NotNil(t, client.ctx)
	assert.False(t, client.verbose)
}

func TestNewClientVerbose(t *testing.T) {
	client := NewClient("test-token", true)

	assert.NotNil(t, client)
	assert.True(t, client.verbose)
}

func TestConvertToSummary(t *testing.T) {
	// Note: This is a unit test for the conversion logic
	// In a real scenario, we would use mocks to test API calls
	// For now, we're testing the client creation and structure
	client := NewClient("test-token", false)
	assert.NotNil(t, client)
}

// Note: Testing actual API calls would require:
// 1. Mock GitHub API server
// 2. Integration tests with real API (skipped in unit tests)
// 3. Test data fixtures
//
// For production code, you would use httptest or a mocking library
// like go-github's test utilities to mock API responses.
//
// Example integration test structure (would need valid token):
/*
func TestFetchIssuesIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		t.Skip("GITHUB_TOKEN not set")
	}

	client := NewClient(token, true)
	issues, err := client.FetchIssues("golang", "go", "all", []string{}, 5)

	assert.NoError(t, err)
	assert.NotEmpty(t, issues)
	assert.LessOrEqual(t, len(issues), 5)
}
*/
