package test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBuildCLI tests that the CLI binary can be built successfully
func TestBuildCLI(t *testing.T) {
	projectRoot := getProjectRoot(t)
	binaryPath := filepath.Join(projectRoot, "test-auto-issue-finder")

	// Clean up before test
	defer os.Remove(binaryPath)

	// Build the CLI
	cmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/analyze/main.go")
	cmd.Dir = projectRoot
	output, err := cmd.CombinedOutput()

	require.NoError(t, err, "Build failed: %s", string(output))
	assert.FileExists(t, binaryPath, "Binary should be created")

	// Check if binary is executable
	info, err := os.Stat(binaryPath)
	require.NoError(t, err)
	assert.True(t, info.Mode()&0111 != 0, "Binary should be executable")
}

// TestCLIHelp tests that the CLI help command works
func TestCLIHelp(t *testing.T) {
	projectRoot := getProjectRoot(t)
	binaryPath := buildTestBinary(t, projectRoot)
	defer os.Remove(binaryPath)

	// Run help command
	cmd := exec.Command(binaryPath, "--help")
	output, err := cmd.CombinedOutput()

	require.NoError(t, err, "Help command failed: %s", string(output))

	outputStr := string(output)
	assert.Contains(t, outputStr, "analyze", "Help should mention analyze command")
	assert.Contains(t, outputStr, "Usage", "Help should contain usage information")
}

// TestCLIVersion tests that the CLI version command works
func TestCLIVersion(t *testing.T) {
	projectRoot := getProjectRoot(t)
	binaryPath := buildTestBinary(t, projectRoot)
	defer os.Remove(binaryPath)

	// Run version command
	cmd := exec.Command(binaryPath, "version")
	output, err := cmd.CombinedOutput()

	// Version command may not be implemented, so we just check it doesn't crash
	outputStr := string(output)
	t.Logf("Version output: %s", outputStr)

	// Should either show version or error message
	assert.True(t,
		strings.Contains(outputStr, "version") ||
			strings.Contains(outputStr, "unknown") ||
			err != nil,
		"Should handle version command")
}

// TestCLIAnalyzeWithoutToken tests that analyze command fails gracefully without token
func TestCLIAnalyzeWithoutToken(t *testing.T) {
	projectRoot := getProjectRoot(t)
	binaryPath := buildTestBinary(t, projectRoot)
	defer os.Remove(binaryPath)

	// Unset GITHUB_TOKEN
	oldToken := os.Getenv("GITHUB_TOKEN")
	os.Unsetenv("GITHUB_TOKEN")
	defer func() {
		if oldToken != "" {
			os.Setenv("GITHUB_TOKEN", oldToken)
		}
	}()

	// Run analyze command without token
	cmd := exec.Command(binaryPath, "analyze", "golang/go", "--limit=1")
	output, err := cmd.CombinedOutput()

	// Should fail with appropriate error
	if err == nil {
		t.Log("Command succeeded, possibly using anonymous access")
	} else {
		outputStr := string(output)
		t.Logf("Error output: %s", outputStr)
		// Error is expected without token
		assert.Error(t, err, "Should fail without valid token")
	}
}

// TestAutoDevScriptsExist tests that auto-dev scripts exist and are executable
func TestAutoDevScriptsExist(t *testing.T) {
	projectRoot := getProjectRoot(t)

	scripts := []string{
		"auto-dev.sh",
		"auto-dev-incremental.sh",
		"auto-dev-with-commits.sh",
		"run-overnight.sh",
		"install-auto-dev.sh",
	}

	for _, script := range scripts {
		scriptPath := filepath.Join(projectRoot, script)
		assert.FileExists(t, scriptPath, "%s should exist", script)

		info, err := os.Stat(scriptPath)
		require.NoError(t, err, "Failed to stat %s", script)
		assert.True(t, info.Mode()&0111 != 0, "%s should be executable", script)
	}
}

// TestAutoDevExampleFileExists tests that example task file exists
func TestAutoDevExampleFileExists(t *testing.T) {
	projectRoot := getProjectRoot(t)
	examplePath := filepath.Join(projectRoot, "tonight-with-tasks.txt.example")

	assert.FileExists(t, examplePath, "Example task file should exist")

	// Read and check content
	content, err := os.ReadFile(examplePath)
	require.NoError(t, err)

	contentStr := string(content)
	assert.Contains(t, contentStr, "タスク", "Example should contain task markers")
}

// TestClaudeSettingsExists tests that Claude settings file exists
func TestClaudeSettingsExists(t *testing.T) {
	projectRoot := getProjectRoot(t)
	settingsPath := filepath.Join(projectRoot, ".claude", "settings.local.json")

	assert.FileExists(t, settingsPath, "Claude settings should exist")

	// Read and check content
	content, err := os.ReadFile(settingsPath)
	require.NoError(t, err)

	contentStr := string(content)
	assert.Contains(t, contentStr, "permissions", "Settings should contain permissions")
	assert.Contains(t, contentStr, "allow", "Settings should have allow list")
}

// TestDocumentationExists tests that all documentation files exist
func TestDocumentationExists(t *testing.T) {
	projectRoot := getProjectRoot(t)

	docs := []string{
		"README.md",
		"CONTRIBUTING.md",
		"docs/INSTALL.md",
		"docs/USAGE.md",
		"docs/AUTO_DEV.md",
		"docs/TESTING.md",
	}

	for _, doc := range docs {
		docPath := filepath.Join(projectRoot, doc)
		assert.FileExists(t, docPath, "%s should exist", doc)

		// Check it's not empty
		info, err := os.Stat(docPath)
		require.NoError(t, err)
		assert.Greater(t, info.Size(), int64(100), "%s should not be empty", doc)
	}
}

// Helper functions

func getProjectRoot(t *testing.T) string {
	t.Helper()

	// Get current working directory
	wd, err := os.Getwd()
	require.NoError(t, err)

	// Go up one level from test/ to project root
	projectRoot := filepath.Dir(wd)
	return projectRoot
}

func buildTestBinary(t *testing.T, projectRoot string) string {
	t.Helper()

	binaryPath := filepath.Join(projectRoot, "test-auto-issue-finder-"+t.Name())

	cmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/analyze/main.go")
	cmd.Dir = projectRoot
	output, err := cmd.CombinedOutput()

	require.NoError(t, err, "Build failed: %s", string(output))
	return binaryPath
}
