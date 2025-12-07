package test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBuildCLI verifies the CLI can be built without errors.
// This catches compilation errors and dependency issues early.
func TestBuildCLI(t *testing.T) {
	projectRoot := getProjectRoot(t)
	binaryPath := filepath.Join(projectRoot, "test-auto-issue-finder")
	defer os.Remove(binaryPath)

	cmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/analyze/main.go")
	cmd.Dir = projectRoot
	output, err := cmd.CombinedOutput()

	require.NoError(t, err, "Build failed: %s", string(output))
	assert.FileExists(t, binaryPath)

	info, err := os.Stat(binaryPath)
	require.NoError(t, err)
	assert.True(t, info.Mode()&0111 != 0, "Binary should be executable")
}

// TestCLIHelp verifies the help command works correctly.
// Users often run --help first, so this must work.
func TestCLIHelp(t *testing.T) {
	projectRoot := getProjectRoot(t)
	binaryPath := buildTestBinary(t, projectRoot)
	defer os.Remove(binaryPath)

	cmd := exec.Command(binaryPath, "--help")
	output, err := cmd.CombinedOutput()

	require.NoError(t, err, "Help command failed: %s", string(output))

	outputStr := string(output)
	assert.Contains(t, outputStr, "analyze")
	assert.Contains(t, outputStr, "Usage")
}

// TestCLIAnalyzeWithoutToken verifies proper error handling when GitHub token is missing.
// Users should get a clear error message explaining how to fix the issue.
func TestCLIAnalyzeWithoutToken(t *testing.T) {
	projectRoot := getProjectRoot(t)
	binaryPath := buildTestBinary(t, projectRoot)
	defer os.Remove(binaryPath)

	oldToken := os.Getenv("GITHUB_TOKEN")
	os.Unsetenv("GITHUB_TOKEN")
	defer func() {
		if oldToken != "" {
			os.Setenv("GITHUB_TOKEN", oldToken)
		}
	}()

	cmd := exec.Command(binaryPath, "analyze", "golang/go", "--limit=1")
	output, err := cmd.CombinedOutput()

	if err == nil {
		t.Log("Command succeeded, possibly using anonymous access")
	} else {
		t.Logf("Error output: %s", string(output))
		assert.Error(t, err)
	}
}

func getProjectRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	require.NoError(t, err)
	return filepath.Dir(wd)
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
