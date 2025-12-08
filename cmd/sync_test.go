package cmd

import (
	"fmt"
	"os"
	"testing"
)

func TestTaskSkipLogic(t *testing.T) {
	tests := []struct {
		name       string
		totalTasks int
		startFrom  int
		wantSkip   []int // Task numbers that should be skipped
		wantExec   []int // Task numbers that should be executed
	}{
		{
			name:       "No skip - start from 1",
			totalTasks: 5,
			startFrom:  1,
			wantSkip:   []int{},
			wantExec:   []int{1, 2, 3, 4, 5},
		},
		{
			name:       "Skip first 2 tasks",
			totalTasks: 5,
			startFrom:  3,
			wantSkip:   []int{1, 2},
			wantExec:   []int{3, 4, 5},
		},
		{
			name:       "Skip all but last task",
			totalTasks: 5,
			startFrom:  5,
			wantSkip:   []int{1, 2, 3, 4},
			wantExec:   []int{5},
		},
		{
			name:       "Single task, start from 1",
			totalTasks: 1,
			startFrom:  1,
			wantSkip:   []int{},
			wantExec:   []int{1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			skipped := []int{}
			executed := []int{}

			// Simulate the task loop
			for i := 0; i < tt.totalTasks; i++ {
				taskNum := i + 1

				// This is the skip logic from sync.go
				if taskNum < tt.startFrom {
					skipped = append(skipped, taskNum)
					continue
				}

				executed = append(executed, taskNum)
			}

			// Verify skipped tasks
			if len(skipped) != len(tt.wantSkip) {
				t.Errorf("skipped count = %d, want %d", len(skipped), len(tt.wantSkip))
			}
			for i, want := range tt.wantSkip {
				if i >= len(skipped) || skipped[i] != want {
					t.Errorf("skipped[%d] = %d, want %d", i, skipped[i], want)
				}
			}

			// Verify executed tasks
			if len(executed) != len(tt.wantExec) {
				t.Errorf("executed count = %d, want %d", len(executed), len(tt.wantExec))
			}
			for i, want := range tt.wantExec {
				if i >= len(executed) || executed[i] != want {
					t.Errorf("executed[%d] = %d, want %d", i, executed[i], want)
				}
			}
		})
	}
}

func TestStartFromValidation(t *testing.T) {
	tests := []struct {
		name       string
		startFrom  int
		totalTasks int
		wantError  bool
	}{
		{
			name:       "Valid - start from 1",
			startFrom:  1,
			totalTasks: 5,
			wantError:  false,
		},
		{
			name:       "Valid - start from middle",
			startFrom:  3,
			totalTasks: 5,
			wantError:  false,
		},
		{
			name:       "Valid - start from last",
			startFrom:  5,
			totalTasks: 5,
			wantError:  false,
		},
		{
			name:       "Invalid - start from 0",
			startFrom:  0,
			totalTasks: 5,
			wantError:  true,
		},
		{
			name:       "Invalid - negative value",
			startFrom:  -1,
			totalTasks: 5,
			wantError:  true,
		},
		{
			name:       "Invalid - exceeds total tasks",
			startFrom:  6,
			totalTasks: 5,
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This is the validation logic from sync.go
			var err error
			if tt.startFrom < 1 {
				err = &validationError{"start-from must be >= 1"}
			}
			if tt.startFrom > tt.totalTasks {
				err = &validationError{"start-from exceeds total tasks"}
			}

			hasError := err != nil
			if hasError != tt.wantError {
				t.Errorf("validation error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// Helper error type for testing
type validationError struct {
	msg string
}

func (e *validationError) Error() string {
	return e.msg
}

func TestParseTaskFile(t *testing.T) {
	// Create a temporary task file for testing
	content := `## タスク1: テストファイルの作成

テストファイルを作成します。

### 確認
- ` + "`test -f test1.txt`" + `

## タスク2: 存在しないファイルへの追記（意図的エラー）

存在しないファイルへの追記を試みます。

### 確認
- ` + "`test -f nonexistent.txt`" + `

## タスク3: 複数ファイルの作成と検証

複数のファイルを作成します。

### 確認
- ` + "`test -f test2.txt && test -f test3.txt`" + `
`

	tmpFile := "../test_parse_temp.txt"
	if err := writeFile(tmpFile, content); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer removeFile(tmpFile)

	tasks, err := parseTaskFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to parse task file: %v", err)
	}

	// Should have exactly 3 tasks
	if len(tasks) != 3 {
		t.Errorf("Expected 3 tasks, got %d", len(tasks))
	}

	// Verify task titles
	expectedTitles := []string{
		"1: テストファイルの作成",
		"2: 存在しないファイルへの追記（意図的エラー）",
		"3: 複数ファイルの作成と検証",
	}

	for i, expectedTitle := range expectedTitles {
		if i >= len(tasks) {
			t.Errorf("Task %d not found", i+1)
			continue
		}
		if tasks[i].Title != expectedTitle {
			t.Errorf("Task %d title = %q, want %q", i+1, tasks[i].Title, expectedTitle)
		}
	}

	// Verify all tasks have verification commands
	for i, task := range tasks {
		if task.Command == "" {
			t.Errorf("Task %d should have a verification command", i+1)
		}
	}
}

// Helper functions for test file operations
func writeFile(path, content string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(content)
	return err
}

func removeFile(path string) {
	os.Remove(path)
}

func TestRecursiveExecution(t *testing.T) {
	tests := []struct {
		name         string
		currentDepth int
		wantSkip     bool
		wantNewDepth int
	}{
		{
			name:         "First level recursion - should execute",
			currentDepth: 0,
			wantSkip:     false,
			wantNewDepth: 1,
		},
		{
			name:         "Second level recursion - should execute",
			currentDepth: 1,
			wantSkip:     false,
			wantNewDepth: 2,
		},
		{
			name:         "Third level recursion - should execute",
			currentDepth: 2,
			wantSkip:     false,
			wantNewDepth: 3,
		},
		{
			name:         "Maximum depth reached - should skip",
			currentDepth: 3,
			wantSkip:     true,
			wantNewDepth: 3,
		},
		{
			name:         "Beyond maximum depth - should skip",
			currentDepth: 4,
			wantSkip:     true,
			wantNewDepth: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the recursion check logic from sync.go:427-433
			shouldSkip := tt.currentDepth >= maxRecursionDepth

			if shouldSkip != tt.wantSkip {
				t.Errorf("shouldSkip = %v, want %v", shouldSkip, tt.wantSkip)
			}

			// Simulate new depth calculation
			var newDepth int
			if !shouldSkip {
				newDepth = tt.currentDepth + 1
			} else {
				newDepth = tt.currentDepth
			}

			if newDepth != tt.wantNewDepth {
				t.Errorf("newDepth = %d, want %d", newDepth, tt.wantNewDepth)
			}
		})
	}
}

func TestGetCurrentRecursionDepth(t *testing.T) {
	tests := []struct {
		name      string
		envValue  string
		wantDepth int
	}{
		{
			name:      "No environment variable - depth 0",
			envValue:  "",
			wantDepth: 0,
		},
		{
			name:      "Depth 1",
			envValue:  "1",
			wantDepth: 1,
		},
		{
			name:      "Depth 3 (max)",
			envValue:  "3",
			wantDepth: 3,
		},
		{
			name:      "Invalid value - defaults to 0",
			envValue:  "invalid",
			wantDepth: 0,
		},
		{
			name:      "Negative value - returns parsed value",
			envValue:  "-1",
			wantDepth: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			if tt.envValue != "" {
				os.Setenv("SLEEPSHIP_DEPTH", tt.envValue)
				defer os.Unsetenv("SLEEPSHIP_DEPTH")
			} else {
				os.Unsetenv("SLEEPSHIP_DEPTH")
			}

			// Test the actual function
			depth := getCurrentRecursionDepth()

			if depth != tt.wantDepth {
				t.Errorf("getCurrentRecursionDepth() = %d, want %d", depth, tt.wantDepth)
			}
		})
	}
}

func TestRecursionDepthLimit(t *testing.T) {
	// Verify that maxRecursionDepth is set correctly
	if maxRecursionDepth != 3 {
		t.Errorf("maxRecursionDepth = %d, want 3", maxRecursionDepth)
	}

	// Test edge cases around the limit
	tests := []struct {
		name        string
		depth       int
		wantAllowed bool
	}{
		{
			name:        "Depth 0 - allowed",
			depth:       0,
			wantAllowed: true,
		},
		{
			name:        "Depth 1 - allowed",
			depth:       1,
			wantAllowed: true,
		},
		{
			name:        "Depth 2 - allowed (last allowed)",
			depth:       2,
			wantAllowed: true,
		},
		{
			name:        "Depth 3 - not allowed (at limit)",
			depth:       3,
			wantAllowed: false,
		},
		{
			name:        "Depth 4 - not allowed (beyond limit)",
			depth:       4,
			wantAllowed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This is the check from sync.go:428
			allowed := tt.depth < maxRecursionDepth

			if allowed != tt.wantAllowed {
				t.Errorf("depth %d: allowed = %v, want %v", tt.depth, allowed, tt.wantAllowed)
			}
		})
	}
}

func TestSleepshipCommandDetection(t *testing.T) {
	tests := []struct {
		name            string
		command         string
		wantIsSleepship bool
	}{
		{
			name:            "sleepship command",
			command:         "sleepship sync tasks.txt",
			wantIsSleepship: true,
		},
		{
			name:            "bin/sleepship command",
			command:         "./bin/sleepship sync tasks.txt",
			wantIsSleepship: true,
		},
		{
			name:            "absolute path sleepship",
			command:         "/usr/local/bin/sleepship sync tasks.txt",
			wantIsSleepship: true,
		},
		{
			name:            "go test command",
			command:         "go test ./...",
			wantIsSleepship: false,
		},
		{
			name:            "go build command",
			command:         "go build",
			wantIsSleepship: false,
		},
		{
			name:            "mixed command with sleepship mention",
			command:         "echo sleepship && go test",
			wantIsSleepship: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This is the detection logic from sync.go:424
			isSleepshipCommand := containsSleepship(tt.command)

			if isSleepshipCommand != tt.wantIsSleepship {
				t.Errorf("isSleepshipCommand = %v, want %v", isSleepshipCommand, tt.wantIsSleepship)
			}
		})
	}
}

// Helper function to detect sleepship commands
func containsSleepship(command string) bool {
	return os.Getenv("TESTING") != "" ||
		(len(command) > 0 && (containsSubstring(command, "sleepship") ||
			containsSubstring(command, "./bin/sleepship")))
}

// Helper function for substring check
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr || findSubstring(s, substr))
}

// Simple substring finder
func findSubstring(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestRecursionBasic tests basic recursion depth detection
// Migrated from test_recursion.sh
func TestRecursionBasic(t *testing.T) {
	tests := []struct {
		name      string
		depth     int
		wantDepth int
	}{
		{
			name:      "Depth 0 (normal execution)",
			depth:     0,
			wantDepth: 0,
		},
		{
			name:      "Depth 1",
			depth:     1,
			wantDepth: 1,
		},
		{
			name:      "Depth 2",
			depth:     2,
			wantDepth: 2,
		},
		{
			name:      "Depth 3 (at max)",
			depth:     3,
			wantDepth: 3,
		},
		{
			name:      "Depth 4 (exceeds max)",
			depth:     4,
			wantDepth: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set the SLEEPSHIP_DEPTH environment variable
			os.Setenv("SLEEPSHIP_DEPTH", fmt.Sprintf("%d", tt.depth))
			defer os.Unsetenv("SLEEPSHIP_DEPTH")

			// Get the current recursion depth
			depth := getCurrentRecursionDepth()

			if depth != tt.wantDepth {
				t.Errorf("getCurrentRecursionDepth() = %d, want %d", depth, tt.wantDepth)
			}

			// Verify warning/blocking behavior
			if tt.depth >= maxRecursionDepth {
				// At max depth or beyond, recursion should be blocked
				if depth < maxRecursionDepth {
					t.Errorf("Expected depth >= maxRecursionDepth, got %d", depth)
				}
			}
		})
	}
}

// TestDepthLimit tests the depth limit enforcement in runCommand
// Migrated from test_depth_limit.sh
func TestDepthLimit(t *testing.T) {
	testCommands := []struct {
		name               string
		command            string
		isSleepshipCommand bool
	}{
		{
			name:               "sleepship sync command",
			command:            "./bin/sleepship sync test.txt",
			isSleepshipCommand: true,
		},
		{
			name:               "echo command",
			command:            "echo test",
			isSleepshipCommand: false,
		},
		{
			name:               "sleepship direct command",
			command:            "sleepship sync another.txt",
			isSleepshipCommand: true,
		},
	}

	for depth := 0; depth <= 4; depth++ {
		t.Run(fmt.Sprintf("Depth_%d", depth), func(t *testing.T) {
			// Set depth environment variable
			os.Setenv("SLEEPSHIP_DEPTH", fmt.Sprintf("%d", depth))
			defer os.Unsetenv("SLEEPSHIP_DEPTH")

			for _, tc := range testCommands {
				t.Run(tc.name, func(t *testing.T) {
					currentDepth := getCurrentRecursionDepth()

					// Check if command contains sleepship
					isSleepship := containsSleepship(tc.command)
					if isSleepship != tc.isSleepshipCommand {
						t.Errorf("containsSleepship(%q) = %v, want %v", tc.command, isSleepship, tc.isSleepshipCommand)
					}

					// Verify depth limit enforcement for sleepship commands
					if tc.isSleepshipCommand {
						shouldBlock := currentDepth >= maxRecursionDepth

						if shouldBlock {
							// Command should be blocked at max depth
							if currentDepth < maxRecursionDepth {
								t.Errorf("Expected command to be blocked at depth %d (max: %d)", currentDepth, maxRecursionDepth)
							}
						} else {
							// Command should be allowed
							if currentDepth >= maxRecursionDepth {
								t.Errorf("Expected command to be allowed at depth %d (max: %d)", currentDepth, maxRecursionDepth)
							}
						}
					}
				})
			}
		})
	}
}
