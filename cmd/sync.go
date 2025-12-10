// Package cmd implements the CLI commands for Sleepship.
package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/isiidaisuke0926/sleepship/internal/config"
	"github.com/isiidaisuke0926/sleepship/internal/history"
	"github.com/spf13/cobra"
)

const (
	maxRecursionDepth = 3 // Maximum depth for recursive sleepship calls
)

var (
	projectDir string
	logDir     string
	worker     bool // Internal flag for background worker process
	startFrom  int  // Start from specified task number
	maxRetries int  // Maximum number of retries for failed verifications (default: 3)
)

// Task represents a development task with title, description, and verification command.
type Task struct {
	Title       string
	Description string
	Command     string // 確認コマンド（go build, go test等）
}

var syncCmd = &cobra.Command{
	Use:   "sync [task-file]",
	Short: "Synchronously execute development tasks with Claude Code",
	Long: "Execute development tasks from a markdown file synchronously with Claude Code.\n\n" +
		"This command reads a task file containing multiple development tasks and executes them\n" +
		"one by one using Claude Code. After each task, it runs verification commands (if specified)\n" +
		"and attempts to fix any errors automatically.\n\n" +
		"Task File Format:\n" +
		"  Tasks are defined using markdown headers starting with \"## タスク\" or \"## Task\".\n" +
		"  Each task can have:\n" +
		"  - Implementation instructions in the body\n" +
		"  - Verification command in a code block starting with \"- `\"\n\n" +
		"Example:\n" +
		"  ## タスク1: Add new feature\n\n" +
		"  ### 実装\n" +
		"  Implement the new feature in src/main.go\n\n" +
		"  ### 確認\n" +
		"  - `go build`\n" +
		"  - `go test ./...`\n\n" +
		"Examples:\n" +
		"  sleepship sync tasks.txt\n" +
		"  sleepship sync tasks.txt --dir=/path/to/project\n" +
		"  sleepship sync tasks.txt --dir=/path/to/project --log-dir=./logs",
	Args: cobra.ExactArgs(1),
	RunE: runSync,
}

func init() {
	rootCmd.AddCommand(syncCmd)

	syncCmd.Flags().StringVar(&projectDir, "dir", "", "Project directory (default: current directory)")
	syncCmd.Flags().StringVar(&logDir, "log-dir", "logs", "Log output directory")
	syncCmd.Flags().IntVar(&startFrom, "start-from", 1, "Start from specified task number (default: 1)")
	syncCmd.Flags().IntVar(&maxRetries, "max-retries", 3, "Maximum number of retries for failed verifications (default: 3)")
	syncCmd.Flags().BoolVar(&worker, "worker", false, "Internal: run as background worker")
	_ = syncCmd.Flags().MarkHidden("worker")
}

//nolint:gocyclo // runSync is complex by nature, handling the full task execution lifecycle
func runSync(cmd *cobra.Command, args []string) error {
	taskFile := args[0]
	startTime := time.Now()

	// Load configuration from environment variables
	envConfig := config.LoadFromEnv()
	defaultConfig := config.NewDefaultConfig()

	// Create CLI config from flags
	cliConfig := &config.Config{
		ProjectDir: projectDir,
		LogDir:     logDir,
		MaxRetries: -1,
		StartFrom:  -1,
	}

	// Check if flags were explicitly set
	if cmd.Flags().Changed("max-retries") {
		cliConfig.MaxRetries = maxRetries
	}
	if cmd.Flags().Changed("start-from") {
		cliConfig.StartFrom = startFrom
	}

	// Merge configurations: CLI > Env > Default
	mergedConfig := config.MergeConfig(cliConfig, config.FromEnv(envConfig), defaultConfig)

	// Apply merged configuration
	projectDir = mergedConfig.ProjectDir
	logDir = mergedConfig.LogDir
	maxRetries = mergedConfig.MaxRetries
	startFrom = mergedConfig.StartFrom

	// Log configuration source for debugging
	if envConfig.HasMaxRetries() && !cmd.Flags().Changed("max-retries") {
		log.Printf("ℹ️  Using max-retries from environment: %d\n", maxRetries)
	}
	if envConfig.HasStartFrom() && !cmd.Flags().Changed("start-from") {
		log.Printf("ℹ️  Using start-from from environment: %d\n", startFrom)
	}
	if envConfig.HasLogDir() && !cmd.Flags().Changed("log-dir") {
		log.Printf("ℹ️  Using log-dir from environment: %s\n", logDir)
	}
	if envConfig.HasProjectDir() && !cmd.Flags().Changed("dir") {
		log.Printf("ℹ️  Using project directory from environment: %s\n", projectDir)
	}

	// If not running as worker, spawn background process
	if !worker {
		return spawnBackgroundWorker(taskFile)
	}

	// Check recursion depth
	currentDepth := getCurrentRecursionDepth()
	if currentDepth > 0 {
		log.Printf("🔁 Recursive execution detected (depth: %d/%d)\n", currentDepth, maxRecursionDepth)
	}

	// Set default project directory to current directory if still empty
	if projectDir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
		projectDir = cwd
	}

	// Convert to absolute path
	absProjectDir, err := filepath.Abs(projectDir)
	if err != nil {
		return fmt.Errorf("failed to resolve project directory: %w", err)
	}
	projectDir = absProjectDir

	// Parse task file
	tasks, err := parseTaskFile(taskFile)
	if err != nil {
		return fmt.Errorf("failed to parse task file: %w", err)
	}

	if len(tasks) == 0 {
		return fmt.Errorf("no tasks found in task file")
	}

	// Validate startFrom value
	if startFrom < 1 {
		return fmt.Errorf("Error: --start-from must be >= 1")
	}
	if startFrom > len(tasks) {
		log.Printf("⚠️ Warning: --start-from (%d) exceeds total tasks (%d). No tasks will be executed.\n", startFrom, len(tasks))
		// Return success but skip all tasks
		fmt.Printf("📋 Total tasks: %d\n", len(tasks))
		fmt.Printf("⏩ Starting from task: %d (exceeds total, nothing to do)\n\n", startFrom)
		fmt.Printf("========================================\n")
		fmt.Printf("✅ All tasks completed successfully!\n")
		fmt.Printf("========================================\n")
		return nil
	}

	// Create log directory
	absLogDir := filepath.Join(projectDir, logDir)
	if err := os.MkdirAll(absLogDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Create log file
	logFileName := fmt.Sprintf("sync-%s.log", time.Now().Format("20060102-150405"))
	logFilePath := filepath.Join(absLogDir, logFileName)
	f, err := os.Create(logFilePath)
	if err != nil {
		return fmt.Errorf("failed to create log file: %w", err)
	}
	defer func() { _ = f.Close() }()

	// Write task info to both stdout and log file
	taskInfo := fmt.Sprintf("📋 Total tasks: %d\n", len(tasks))
	fmt.Print(taskInfo)
	_, _ = f.WriteString(taskInfo)

	if startFrom > 1 {
		startInfo := fmt.Sprintf("⏩ Starting from task: %d\n", startFrom)
		fmt.Print(startInfo)
		_, _ = f.WriteString(startInfo)
	}

	dirInfo := fmt.Sprintf("📁 Project directory: %s\n\n", projectDir)
	fmt.Print(dirInfo)
	_, _ = f.WriteString(dirInfo)

	// Create branch for this sync execution
	var branchName string
	if err := createBranchForSync(taskFile, f); err != nil {
		log.Printf("⚠️ Warning: Failed to create branch: %v\n", err)
		// Continue anyway - branch creation is not critical
		branchName = ""
	} else {
		// Extract branch name from task file
		filename := filepath.Base(taskFile)
		sanitized := sanitizeBranchName(filename)
		branchName = fmt.Sprintf("feature/%s", sanitized)
	}

	// Execute tasks
	for i, task := range tasks {
		taskNum := i + 1

		// Skip tasks before startFrom
		if taskNum < startFrom {
			skipMsg := fmt.Sprintf("⏭️  Skipping task %d/%d (start-from=%d): %s\n", taskNum, len(tasks), startFrom, task.Title)
			fmt.Print(skipMsg)
			_, _ = f.WriteString(skipMsg)
			continue
		}

		taskHeader := fmt.Sprintf("========================================\nTask %d/%d: %s\n========================================\n\n", taskNum, len(tasks), task.Title)
		fmt.Print(taskHeader)
		_, _ = f.WriteString(taskHeader)

		// Execute task with Claude with retry logic
		var lastErr error
		taskRetryCount := 0

		for taskRetryCount <= maxRetries {
			if err := executeTask(task, f); err != nil {
				lastErr = err
				taskRetryCount++

				// Detailed error logging
				errorMsg := fmt.Sprintf("Task failed (task %d/%d, attempt %d/%d): %v", taskNum, len(tasks), taskRetryCount, maxRetries+1, err)
				_, _ = fmt.Fprintf(f, "\n❌ %s\n", errorMsg)

				if taskRetryCount > maxRetries {
					// Final failure - no more retries available
					failureMsg := fmt.Sprintf("❌ Task failed: タスク %d (\"%s\") が %d 回の試行後も失敗しました\n", taskNum, task.Title, maxRetries+1)
					fmt.Print(failureMsg)
					_, _ = f.WriteString(failureMsg)

					detailedError := fmt.Sprintf("【失敗の詳細】\n")
					detailedError += fmt.Sprintf("  タスク番号: %d/%d\n", taskNum, len(tasks))
					detailedError += fmt.Sprintf("  タスク名: %s\n", task.Title)
					detailedError += fmt.Sprintf("  試行回数: %d回\n", maxRetries+1)
					detailedError += fmt.Sprintf("  エラー内容: %v\n", err)
					detailedError += fmt.Sprintf("\n実行を停止します。リトライ不可。\n")
					fmt.Print(detailedError)
					_, _ = f.WriteString(detailedError)

					log.Printf("❌ Task failed: リトライ上限に達しました (task %d, max retries: %d)\n", taskNum, maxRetries)

					// Record failed execution to history
					duration := time.Since(startTime)
					histErr := history.Record(projectDir, taskFile, branchName, false, duration, len(tasks), startFrom, maxRetries, fmt.Sprintf("Task %d failed: %v", taskNum, err))
					if histErr != nil {
						log.Printf("⚠️ Warning: Failed to record history: %v\n", histErr)
					}

					return fmt.Errorf("task %d failed after %d attempts: %w", taskNum, maxRetries+1, err)
				}

				// Retry is possible
				retryMsg := fmt.Sprintf("❌ タスク %d の実行に失敗しました。リトライ可能: %d/%d 回目を実行します\n", taskNum, taskRetryCount, maxRetries)
				fmt.Print(retryMsg)
				_, _ = f.WriteString(retryMsg)

				retryDetail := fmt.Sprintf("【リトライ詳細】\n")
				retryDetail += fmt.Sprintf("  タスク番号: %d/%d\n", taskNum, len(tasks))
				retryDetail += fmt.Sprintf("  タスク名: %s\n", task.Title)
				retryDetail += fmt.Sprintf("  現在の試行: %d回目\n", taskRetryCount)
				retryDetail += fmt.Sprintf("  残りリトライ: %d回\n", maxRetries-taskRetryCount+1)
				retryDetail += fmt.Sprintf("  エラー内容: %v\n", err)
				fmt.Print(retryDetail)
				_, _ = f.WriteString(retryDetail)

				log.Printf("🔄 Task retry: リトライを開始します (task %d, attempt %d/%d)\n", taskNum, taskRetryCount, maxRetries)

				// Retry with error context
				retryPrompt := fmt.Sprintf(`前回のタスク実行でエラーが発生しました (リトライ %d/%d):
エラー: %v

# タスク
%s

%s

# 指示
1. 前回のエラーを修正してください
2. このタスクを完全に実装してください
3. 必要なファイルを作成・編集してください
4. 実装後、必ず動作確認してください
5. エラーがあれば修正してください

プロジェクトディレクトリ: %s

実装を開始してください。

実装後、以下の質問に必ず答えてください：

【成功判定】
このタスクは完全に成功しましたか？以下を確認してください：
1. 必要なファイルが作成されているか
2. 確認コマンドが成功しているか
3. エラーが発生していないか

成功の場合: "SUCCESS: このタスクは成功しました"
失敗の場合: "FAILED: このタスクは失敗しました。理由: [具体的な理由]"

という形式で必ず応答してください。`, taskRetryCount, maxRetries, err, task.Title, task.Description, projectDir)

				if err := executeClaude(retryPrompt, f); err != nil {
					log.Printf("❌ リトライ実行に失敗しました: %v\n", err)
					// Continue to next retry attempt
					continue
				}

				// Retry succeeded, break out of retry loop
				lastErr = nil
				break
			}

			// Task succeeded on first try
			break
		}

		if lastErr != nil {
			return fmt.Errorf("task %d failed after all retries: %w", taskNum, lastErr)
		}

		if taskRetryCount > 0 {
			fmt.Printf("✅ タスク %d が %d 回のリトライ後に成功しました\n", taskNum, taskRetryCount)
		}

		// Run verification command with retry logic
		if task.Command != "" {
			fmt.Printf("\n🔍 Running verification: %s\n", task.Command)

			retryCount := 0
			verificationPassed := false

			for retryCount <= maxRetries {
				if err := runCommand(task.Command, f); err != nil {
					retryCount++

					// Detailed verification error logging
					verifyErrorMsg := fmt.Sprintf("Task failed (verification): task %d/%d, command: %s, attempt %d/%d", taskNum, len(tasks), task.Command, retryCount, maxRetries+1)
					_, _ = fmt.Fprintf(f, "\n❌ %s\n", verifyErrorMsg)

					if retryCount > maxRetries {
						// Final verification failure - no more retries
						verifyFailMsg := fmt.Sprintf("❌ Task failed (verification): タスク %d (\"%s\") の検証が %d 回の試行後も失敗しました\n", taskNum, task.Title, maxRetries+1)
						fmt.Print(verifyFailMsg)
						_, _ = f.WriteString(verifyFailMsg)

						detailedVerifyError := fmt.Sprintf("【検証失敗の詳細】\n")
						detailedVerifyError += fmt.Sprintf("  タスク番号: %d/%d\n", taskNum, len(tasks))
						detailedVerifyError += fmt.Sprintf("  タスク名: %s\n", task.Title)
						detailedVerifyError += fmt.Sprintf("  検証コマンド: %s\n", task.Command)
						detailedVerifyError += fmt.Sprintf("  試行回数: %d回\n", maxRetries+1)
						detailedVerifyError += fmt.Sprintf("  エラー内容: %v\n", err)
						detailedVerifyError += fmt.Sprintf("\n実行を停止します。リトライ不可。\n")
						fmt.Print(detailedVerifyError)
						_, _ = f.WriteString(detailedVerifyError)

						log.Printf("❌ Task failed (verification): リトライ上限に達しました (task %d, max retries: %d)\n", taskNum, maxRetries)

						// Record failed execution to history
						duration := time.Since(startTime)
						histErr := history.Record(projectDir, taskFile, branchName, false, duration, len(tasks), startFrom, maxRetries, fmt.Sprintf("Verification failed for task %d: %v", taskNum, err))
						if histErr != nil {
							log.Printf("⚠️ Warning: Failed to record history: %v\n", histErr)
						}

						return fmt.Errorf("verification failed after %d attempts: %w", maxRetries+1, err)
					}

					// Verification retry is possible
					verifyRetryMsg := fmt.Sprintf("❌ 検証失敗、修正を試みます（リトライ可能: %d/%d 回目）\n", retryCount, maxRetries)
					fmt.Print(verifyRetryMsg)
					_, _ = f.WriteString(verifyRetryMsg)

					verifyRetryDetail := fmt.Sprintf("【検証リトライ詳細】\n")
					verifyRetryDetail += fmt.Sprintf("  タスク番号: %d/%d\n", taskNum, len(tasks))
					verifyRetryDetail += fmt.Sprintf("  タスク名: %s\n", task.Title)
					verifyRetryDetail += fmt.Sprintf("  検証コマンド: %s\n", task.Command)
					verifyRetryDetail += fmt.Sprintf("  現在の試行: %d回目\n", retryCount)
					verifyRetryDetail += fmt.Sprintf("  残りリトライ: %d回\n", maxRetries-retryCount+1)
					verifyRetryDetail += fmt.Sprintf("  エラー内容: %v\n", err)
					fmt.Print(verifyRetryDetail)
					_, _ = f.WriteString(verifyRetryDetail)

					log.Printf("🔄 Task retry (verification): リトライを開始します (task %d, attempt %d/%d)\n", taskNum, retryCount, maxRetries)

					// Attempt to fix
					fixPrompt := fmt.Sprintf(`検証コマンドが失敗しました（リトライ %d/%d 回目）:

コマンド: %s
エラー: %v

# 指示
1. 上記のエラーを修正してください
2. 修正後、検証が通ることを確認してください
3. エラーがあれば修正してください

プロジェクトディレクトリ: %s

修正を開始してください。`, retryCount, maxRetries, task.Command, err, projectDir)

					if err := executeClaude(fixPrompt, f); err != nil {
						log.Printf("❌ 修正の実行に失敗しました: %v\n", err)
						// Continue to next retry attempt
						continue
					}

					log.Printf("🔍 修正後、検証を再実行します...\n")
					// Continue to retry verification
					continue
				}

				// Verification passed
				verificationPassed = true
				if retryCount > 0 {
					fmt.Printf("✅ 検証が %d 回のリトライ後に成功しました\n", retryCount)
				} else {
					fmt.Printf("✅ Verification passed\n")
				}
				break
			}

			if !verificationPassed {
				return fmt.Errorf("verification failed after all retries")
			}
		}

		// Commit changes for this task
		if err := commitTaskChanges(task, taskNum, f); err != nil {
			log.Printf("⚠️ Warning: Failed to commit changes: %v\n", err)
			// Continue anyway - commit failure is not critical
		}

		fmt.Printf("\n✅ Task %d completed\n\n", taskNum)
		time.Sleep(1 * time.Second)
	}

	fmt.Printf("========================================\n")
	fmt.Printf("✅ All tasks completed successfully!\n")
	fmt.Printf("========================================\n")
	fmt.Printf("📝 Log file: %s\n", logFilePath)

	// Record successful execution to history
	duration := time.Since(startTime)
	if err := history.Record(projectDir, taskFile, branchName, true, duration, len(tasks), startFrom, maxRetries, ""); err != nil {
		log.Printf("⚠️ Warning: Failed to record history: %v\n", err)
	}

	// Generate and display PR information
	generatePRInfo(tasks, taskFile)

	return nil
}

func parseTaskFile(filename string) ([]Task, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var tasks []Task
	var currentTask *Task
	var descLines []string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// Task title (starts with "## タスク" or "## Task")
		if strings.HasPrefix(line, "## タスク") || strings.HasPrefix(line, "## Task") {
			// Save previous task
			if currentTask != nil {
				currentTask.Description = strings.Join(descLines, "\n")
				tasks = append(tasks, *currentTask)
			}

			// Start new task
			currentTask = &Task{
				Title: strings.TrimPrefix(strings.TrimPrefix(line, "## タスク"), "## Task"),
			}
			descLines = []string{}
			continue
		}

		// Verification command (line starting with "- `")
		if currentTask != nil && strings.HasPrefix(line, "- `") && strings.HasSuffix(line, "`") {
			// Extract command from "- `command`" format
			cmd := strings.TrimPrefix(line, "- `")
			cmd = strings.TrimSuffix(cmd, "`")
			currentTask.Command = cmd
			continue
		}

		// Accumulate description lines
		if currentTask != nil && line != "" && !strings.HasPrefix(line, "---") {
			descLines = append(descLines, line)
		}
	}

	// Save last task
	if currentTask != nil {
		currentTask.Description = strings.Join(descLines, "\n")
		tasks = append(tasks, *currentTask)
	}

	return tasks, scanner.Err()
}

func executeTask(task Task, logFile *os.File) error {
	prompt := fmt.Sprintf(`あなたは自律的にソフトウェア開発を行うエンジニアです。

# タスク
%s

%s

# 指示
1. このタスクを完全に実装してください
2. 必要なファイルを作成・編集してください
3. 実装後、必ず動作確認してください
4. エラーがあれば修正してください

プロジェクトディレクトリ: %s

実装を開始してください。

実装後、以下の質問に必ず答えてください：

【成功判定】
このタスクは完全に成功しましたか？以下を確認してください：
1. 必要なファイルが作成されているか
2. 確認コマンドが成功しているか
3. エラーが発生していないか

成功の場合: "SUCCESS: このタスクは成功しました"
失敗の場合: "FAILED: このタスクは失敗しました。理由: [具体的な理由]"

という形式で必ず応答してください。`, task.Title, task.Description, projectDir)

	if err := executeClaude(prompt, logFile); err != nil {
		return err
	}

	// 確認コマンドを実行
	if verifyCmd := task.Command; verifyCmd != "" {
		fmt.Printf("\n🔍 Running verification in executeTask: %s\n", verifyCmd)
		_, _ = fmt.Fprintf(logFile, "\n=== Verification Command: %s ===\n", verifyCmd)

		cmd := exec.Command("bash", "-c", verifyCmd)
		cmd.Dir = projectDir
		output, err := cmd.CombinedOutput()
		_, _ = logFile.Write(output)

		if err != nil {
			return fmt.Errorf("verification failed: %s\nOutput: %s", err, string(output))
		}
		log.Printf("✅ Verification passed: %s", verifyCmd)
		_, _ = fmt.Fprintf(logFile, "✅ Verification passed\n")
	}

	return nil
}

func executeClaude(prompt string, logFile *os.File) error {
	cmd := exec.Command("claude", "-p", "--dangerously-skip-permissions")
	cmd.Stdin = strings.NewReader(prompt)
	cmd.Dir = projectDir

	// Output to both screen and log file
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	_, _ = fmt.Fprintf(logFile, "\n=== Claude Execution ===\n%s\n\n", time.Now().Format("2006-01-02 15:04:05"))
	_, _ = logFile.WriteString(prompt)
	_, _ = logFile.WriteString("\n\n")

	fmt.Println("🤖 Executing with Claude...")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("claude execution failed: %w", err)
	}

	return nil
}

func runCommand(command string, logFile *os.File) error {
	_, _ = fmt.Fprintf(logFile, "\n=== Command Execution: %s ===\n", command)

	// Check if command is a sleepship call
	isSleepshipCommand := strings.Contains(command, "sleepship") || strings.Contains(command, "./bin/sleepship")

	if isSleepshipCommand {
		currentDepth := getCurrentRecursionDepth()
		if currentDepth >= maxRecursionDepth {
			warningMsg := fmt.Sprintf("⚠️ Maximum recursion depth (%d) reached. Skipping sleepship command: %s\n", maxRecursionDepth, command)
			fmt.Print(warningMsg)
			_, _ = logFile.WriteString(warningMsg)
			return nil // Don't treat as error, just skip
		}
		log.Printf("🔁 Executing recursive sleepship command (depth: %d -> %d)\n", currentDepth, currentDepth+1)
	}

	cmd := exec.Command("bash", "-c", command)
	cmd.Dir = projectDir

	// Set environment variables for recursive execution
	cmd.Env = os.Environ()
	if isSleepshipCommand {
		currentDepth := getCurrentRecursionDepth()
		cmd.Env = append(cmd.Env, fmt.Sprintf("SLEEPSHIP_DEPTH=%d", currentDepth+1))
	}

	output, err := cmd.CombinedOutput()
	_, _ = logFile.Write(output)

	if err != nil {
		return fmt.Errorf("%w\nOutput: %s", err, string(output))
	}

	fmt.Printf("Output: %s\n", string(output))
	return nil
}

func getCurrentRecursionDepth() int {
	depthStr := os.Getenv("SLEEPSHIP_DEPTH")
	if depthStr == "" {
		return 0
	}
	depth, err := strconv.Atoi(depthStr)
	if err != nil {
		return 0
	}
	return depth
}

func sanitizeBranchName(name string) string {
	// Remove file extension
	name = strings.TrimSuffix(name, ".txt")
	name = strings.TrimSuffix(name, ".md")

	// Remove common prefixes
	name = strings.TrimPrefix(name, "tasks-")
	name = strings.TrimPrefix(name, "task-")

	// Convert camelCase to kebab-case
	re := regexp.MustCompile(`([a-z0-9])([A-Z])`)
	name = re.ReplaceAllString(name, "${1}-${2}")

	// Convert to lowercase
	name = strings.ToLower(name)

	// Remove non-alphanumeric characters (keep hyphens)
	re = regexp.MustCompile(`[^a-z0-9-]+`)
	name = re.ReplaceAllString(name, "-")

	// Replace multiple hyphens with single hyphen
	re = regexp.MustCompile(`-+`)
	name = re.ReplaceAllString(name, "-")

	// Trim hyphens from start and end
	name = strings.Trim(name, "-")

	// Limit length
	if len(name) > 50 {
		name = name[:50]
		name = strings.Trim(name, "-")
	}

	// Fallback if empty
	if name == "" {
		name = "sync-" + time.Now().Format("20060102-150405")
	}

	return name
}

func createBranchForSync(taskFile string, logFile *os.File) error {
	// Extract filename from path
	filename := filepath.Base(taskFile)
	sanitized := sanitizeBranchName(filename)
	branchName := fmt.Sprintf("feature/%s", sanitized)

	fmt.Printf("🌿 Creating branch: %s\n", branchName)
	_, _ = fmt.Fprintf(logFile, "\n=== Creating Branch: %s ===\n", branchName)

	cmd := exec.Command("git", "checkout", "-b", branchName)
	cmd.Dir = projectDir

	output, err := cmd.CombinedOutput()
	_, _ = logFile.Write(output)

	if err != nil {
		return fmt.Errorf("failed to create branch: %w\nOutput: %s", err, string(output))
	}

	fmt.Printf("✅ Branch created: %s\n\n", branchName)
	return nil
}

func commitTaskChanges(task Task, taskNumber int, logFile *os.File) error {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	commitMessage := fmt.Sprintf("タスク%d: %s (%s)", taskNumber, task.Title, timestamp)

	fmt.Printf("\n💾 Committing changes: %s\n", commitMessage)
	_, _ = logFile.WriteString("\n=== Committing Changes ===\n")

	// Add all changes
	addCmd := exec.Command("git", "add", ".")
	addCmd.Dir = projectDir
	addOutput, err := addCmd.CombinedOutput()
	_, _ = logFile.Write(addOutput)

	if err != nil {
		return fmt.Errorf("failed to add changes: %w\nOutput: %s", err, string(addOutput))
	}

	// Commit changes
	commitCmd := exec.Command("git", "commit", "-m", commitMessage)
	commitCmd.Dir = projectDir
	commitOutput, err := commitCmd.CombinedOutput()
	_, _ = logFile.Write(commitOutput)

	if err != nil {
		// Check if there are no changes to commit
		if strings.Contains(string(commitOutput), "nothing to commit") {
			fmt.Printf("ℹ️ No changes to commit\n")
			return nil
		}
		return fmt.Errorf("failed to commit: %w\nOutput: %s", err, string(commitOutput))
	}

	fmt.Printf("✅ Changes committed\n")
	return nil
}

func generatePRInfo(tasks []Task, taskFile string) {
	// Extract feature name from task file
	filename := filepath.Base(taskFile)
	featureName := sanitizeBranchName(filename)

	// Generate PR title
	prTitle := generatePRTitle(tasks, featureName)

	// Generate PR body
	prBody := generatePRBody(tasks)

	// Display PR information
	fmt.Printf("\n========================================\n")
	fmt.Printf("📋 プルリクエスト情報\n")
	fmt.Printf("========================================\n\n")

	fmt.Printf("📌 タイトル:\n%s\n\n", prTitle)
	fmt.Printf("📝 本文:\n%s\n", prBody)
	fmt.Printf("========================================\n")
}

func generatePRTitle(tasks []Task, featureName string) string {
	// Use first task title as base, or use feature name
	if len(tasks) > 0 {
		firstTaskTitle := tasks[0].Title
		// Remove task number prefix
		re := regexp.MustCompile(`^(\d+|タスク\d+|Task\d+):\s*`)
		firstTaskTitle = re.ReplaceAllString(firstTaskTitle, "")

		// If it's a comprehensive feature, use it as title
		if len(tasks) > 5 {
			return fmt.Sprintf("%sの実装", firstTaskTitle)
		}
		return firstTaskTitle
	}

	// Fallback: use feature name
	return fmt.Sprintf("%sの実装", featureName)
}

func generatePRBody(tasks []Task) string {
	var body strings.Builder

	body.WriteString("## 概要\n\n")
	body.WriteString(fmt.Sprintf("このPRでは、以下の%d個のタスクを実装しました。\n\n", len(tasks)))

	body.WriteString("## 実装内容\n\n")
	for i, task := range tasks {
		// Remove task number prefix from title
		re := regexp.MustCompile(`^(\d+|タスク\d+|Task\d+):\s*`)
		taskTitle := re.ReplaceAllString(task.Title, "")

		body.WriteString(fmt.Sprintf("%d. %s\n", i+1, taskTitle))
	}

	body.WriteString("\n## テスト\n\n")
	body.WriteString("各タスク完了時に以下の確認を実施済み:\n\n")

	// Collect unique verification commands
	verificationCmds := make(map[string]bool)
	for _, task := range tasks {
		if task.Command != "" {
			verificationCmds[task.Command] = true
		}
	}

	for cmd := range verificationCmds {
		body.WriteString(fmt.Sprintf("- `%s`\n", cmd))
	}

	body.WriteString("\n## 備考\n\n")
	body.WriteString("このPRは自律開発ツール（sleepship）により自動生成されました。\n")

	return body.String()
}

func spawnBackgroundWorker(taskFile string) error {
	// Get current working directory for project dir
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Use provided projectDir or default to cwd
	targetDir := projectDir
	if targetDir == "" {
		targetDir = cwd
	}

	// Create log directory
	absLogDir := filepath.Join(targetDir, logDir)
	if err := os.MkdirAll(absLogDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Generate log file name
	logFileName := fmt.Sprintf("sync-%s.log", time.Now().Format("20060102-150405"))
	logFilePath := filepath.Join(absLogDir, logFileName)

	// Open log file
	logFile, err := os.Create(logFilePath)
	if err != nil {
		return fmt.Errorf("failed to create log file: %w", err)
	}
	defer func() { _ = logFile.Close() }()

	// Get executable path
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Build command arguments
	cmdArgs := []string{"sync", taskFile, "--worker"}
	if projectDir != "" {
		cmdArgs = append(cmdArgs, "--dir", projectDir)
	}
	if logDir != "logs" {
		cmdArgs = append(cmdArgs, "--log-dir", logDir)
	}
	if startFrom != 1 {
		cmdArgs = append(cmdArgs, "--start-from", fmt.Sprintf("%d", startFrom))
	}
	if maxRetries != 3 {
		cmdArgs = append(cmdArgs, "--max-retries", fmt.Sprintf("%d", maxRetries))
	}

	// Start background process
	cmd := exec.Command(executable, cmdArgs...)
	cmd.Dir = cwd
	cmd.Stdout = logFile
	cmd.Stderr = logFile

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start background process: %w", err)
	}

	// Display status
	fmt.Printf("✅ Started background execution (PID: %d)\n", cmd.Process.Pid)
	fmt.Printf("📝 Log file: %s\n", logFilePath)
	fmt.Printf("💡 Monitor: tail -f %s\n", logFilePath)

	// Don't wait for the process to finish
	return nil
}
