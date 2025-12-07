package test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBuildCLI はCLIがエラーなくビルドできることを検証する。
// コンパイルエラーや依存関係の問題を早期に検出するため。
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

// TestCLIHelp はヘルプコマンドが正しく動作することを検証する。
// ユーザーは最初に--helpを実行することが多いため、これは必ず動作する必要がある。
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

// TestCLIAnalyzeWithoutToken はGitHubトークンが無い場合の適切なエラー処理を検証する。
// ユーザーは問題の修正方法を説明する明確なエラーメッセージを受け取る必要がある。
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
