// Package task provides task parsing and management functionality.
package task

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Task represents a development task with title, description, and verification command.
type Task struct {
	Title         string
	Description   string
	Command       string // 確認コマンド（go build, go test等）
	Prerequisites string // 前提確認コマンド
	Dependencies  []int  // 依存するタスク番号のリスト
}

// ParseTaskFile parses a task file and returns a list of tasks.
func ParseTaskFile(filename string) ([]Task, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var tasks []Task
	var currentTask *Task
	var descLines []string
	var inDependencySection bool
	var inPrerequisiteSection bool

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
				Title:        strings.TrimPrefix(strings.TrimPrefix(line, "## タスク"), "## Task"),
				Dependencies: []int{},
			}
			descLines = []string{}
			inDependencySection = false
			inPrerequisiteSection = false
			continue
		}

		// Dependency section header
		if currentTask != nil && (strings.HasPrefix(line, "### 依存") || strings.HasPrefix(line, "### Dependencies")) {
			inDependencySection = true
			inPrerequisiteSection = false
			continue
		}

		// Prerequisite section header
		if currentTask != nil && (strings.HasPrefix(line, "### 前提確認") || strings.HasPrefix(line, "### Prerequisites")) {
			inPrerequisiteSection = true
			inDependencySection = false
			continue
		}

		// Exit special sections on next section header
		if currentTask != nil && strings.HasPrefix(line, "###") {
			if !strings.HasPrefix(line, "### 依存") &&
				!strings.HasPrefix(line, "### Dependencies") &&
				!strings.HasPrefix(line, "### 前提確認") &&
				!strings.HasPrefix(line, "### Prerequisites") {
				inDependencySection = false
				inPrerequisiteSection = false
			}
		}

		// Parse dependency lines (e.g., "- 1, 2, 3" or "- 1")
		if currentTask != nil && inDependencySection && strings.HasPrefix(line, "- ") {
			depStr := strings.TrimPrefix(line, "- ")
			deps := parseDependencies(depStr)
			currentTask.Dependencies = append(currentTask.Dependencies, deps...)
			continue
		}

		// Parse prerequisite command (line starting with "- `")
		if currentTask != nil && inPrerequisiteSection && strings.HasPrefix(line, "- `") && strings.HasSuffix(line, "`") {
			// Extract command from "- `command`" format
			cmd := strings.TrimPrefix(line, "- `")
			cmd = strings.TrimSuffix(cmd, "`")
			currentTask.Prerequisites = cmd
			continue
		}

		// Verification command (line starting with "- `")
		if currentTask != nil && !inDependencySection && !inPrerequisiteSection && strings.HasPrefix(line, "- `") && strings.HasSuffix(line, "`") {
			// Extract command from "- `command`" format
			cmd := strings.TrimPrefix(line, "- `")
			cmd = strings.TrimSuffix(cmd, "`")
			currentTask.Command = cmd
			continue
		}

		// Accumulate description lines
		if currentTask != nil && !inDependencySection && !inPrerequisiteSection && line != "" && !strings.HasPrefix(line, "---") {
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

// parseDependencies parses a comma-separated list of task numbers
// Example: "1, 2, 3" -> [1, 2, 3]
func parseDependencies(depStr string) []int {
	var deps []int
	parts := strings.Split(depStr, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if num, err := strconv.Atoi(part); err == nil {
			deps = append(deps, num)
		}
	}
	return deps
}

// ValidateDependencies validates that all task dependencies are valid
func ValidateDependencies(tasks []Task) error {
	// Create a map of valid task numbers
	validTasks := make(map[int]bool)
	for i := range tasks {
		validTasks[i+1] = true
	}

	// Check each task's dependencies
	for i, task := range tasks {
		taskNum := i + 1
		for _, dep := range task.Dependencies {
			// Check if dependency exists
			if !validTasks[dep] {
				return fmt.Errorf("task %d references non-existent task %d", taskNum, dep)
			}
			// Check for self-dependency
			if dep == taskNum {
				return fmt.Errorf("task %d cannot depend on itself", taskNum)
			}
			// Check for forward dependency (task cannot depend on later tasks)
			if dep >= taskNum {
				return fmt.Errorf("task %d cannot depend on later task %d (dependencies must be on earlier tasks)", taskNum, dep)
			}
		}
	}

	// Check for circular dependencies (if any exist)
	if err := checkCircularDependencies(tasks); err != nil {
		return err
	}

	return nil
}

// checkCircularDependencies checks for circular dependencies in tasks
func checkCircularDependencies(_ []Task) error {
	// Since we only allow dependencies on earlier tasks (checked above),
	// circular dependencies are impossible in this design.
	// This function is kept for completeness and future extensibility.
	return nil
}
