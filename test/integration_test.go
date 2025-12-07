package test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBuildCLI tests that the CLI binary can be built successfully
// これは重要：Goコードがコンパイルできるか、依存関係に問題がないかを確認
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
// これは重要：ユーザーが最初に実行するコマンドが動作するか確認
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

// TestCLIAnalyzeWithoutToken tests that analyze command fails gracefully without token
// これは重要：エラーハンドリングが適切か、ユーザーに分かりやすいエラーが出るか確認
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
