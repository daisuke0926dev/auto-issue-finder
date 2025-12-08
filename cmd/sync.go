package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	projectDir string
	logDir     string
	worker     bool // Internal flag for background worker process
)

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
		"  auto-issue-finder sync tasks.txt\n" +
		"  auto-issue-finder sync tasks.txt --dir=/path/to/project\n" +
		"  auto-issue-finder sync tasks.txt --dir=/path/to/project --log-dir=./logs",
	Args: cobra.ExactArgs(1),
	RunE: runSync,
}

func init() {
	rootCmd.AddCommand(syncCmd)

	syncCmd.Flags().StringVar(&projectDir, "dir", "", "Project directory (default: current directory)")
	syncCmd.Flags().StringVar(&logDir, "log-dir", "logs", "Log output directory")
	syncCmd.Flags().BoolVar(&worker, "worker", false, "Internal: run as background worker")
	syncCmd.Flags().MarkHidden("worker")
}

func runSync(cmd *cobra.Command, args []string) error {
	taskFile := args[0]

	// If not running as worker, spawn background process
	if !worker {
		return spawnBackgroundWorker(taskFile)
	}

	// Set default project directory to current directory
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

	fmt.Printf("📋 Total tasks: %d\n", len(tasks))
	fmt.Printf("📁 Project directory: %s\n\n", projectDir)

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
	defer f.Close()

	// Execute tasks
	for i, task := range tasks {
		fmt.Printf("========================================\n")
		fmt.Printf("Task %d/%d: %s\n", i+1, len(tasks), task.Title)
		fmt.Printf("========================================\n\n")

		// Execute task with Claude
		if err := executeTask(task, f); err != nil {
			log.Printf("❌ Task %d failed: %v\n", i+1, err)
			log.Printf("Stopping execution.\n")
			return fmt.Errorf("task %d failed: %w", i+1, err)
		}

		// Run verification command
		if task.Command != "" {
			fmt.Printf("\n🔍 Running verification: %s\n", task.Command)
			if err := runCommand(task.Command, f); err != nil {
				log.Printf("❌ Verification failed: %v\n", err)
				log.Printf("Attempting to fix...\n")

				// Attempt to fix
				fixPrompt := fmt.Sprintf("前のタスクでエラーが発生しました:\n%v\n\n修正してください。", err)
				if err := executeClaude(fixPrompt, f); err != nil {
					log.Printf("❌ Fix failed: %v\n", err)
					return fmt.Errorf("fix failed after verification error: %w", err)
				}

				// Retry verification
				if err := runCommand(task.Command, f); err != nil {
					log.Printf("❌ Still failing after fix: %v\n", err)
					return fmt.Errorf("verification still failing after fix: %w", err)
				}
			}
			fmt.Printf("✅ Verification passed\n")
		}

		fmt.Printf("\n✅ Task %d completed\n\n", i+1)
		time.Sleep(1 * time.Second)
	}

	fmt.Printf("========================================\n")
	fmt.Printf("✅ All tasks completed successfully!\n")
	fmt.Printf("========================================\n")
	fmt.Printf("📝 Log file: %s\n", logFilePath)

	return nil
}

func parseTaskFile(filename string) ([]Task, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

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

実装を開始してください。`, task.Title, task.Description, projectDir)

	return executeClaude(prompt, logFile)
}

func executeClaude(prompt string, logFile *os.File) error {
	cmd := exec.Command("claude", "-p", "--dangerously-skip-permissions")
	cmd.Stdin = strings.NewReader(prompt)
	cmd.Dir = projectDir

	// Output to both screen and log file
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	logFile.WriteString(fmt.Sprintf("\n=== Claude Execution ===\n%s\n\n", time.Now().Format("2006-01-02 15:04:05")))
	logFile.WriteString(prompt)
	logFile.WriteString("\n\n")

	fmt.Println("🤖 Executing with Claude...")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("claude execution failed: %w", err)
	}

	return nil
}

func runCommand(command string, logFile *os.File) error {
	logFile.WriteString(fmt.Sprintf("\n=== Command Execution: %s ===\n", command))

	cmd := exec.Command("bash", "-c", command)
	cmd.Dir = projectDir

	output, err := cmd.CombinedOutput()
	logFile.Write(output)

	if err != nil {
		return fmt.Errorf("%s\nOutput: %s", err, string(output))
	}

	fmt.Printf("Output: %s\n", string(output))
	return nil
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
	defer logFile.Close()

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
