// Package history provides task execution history tracking for Sleepship.
package history

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	historyDir  = ".sleepship"
	historyFile = "history.json"
)

// Entry represents a single task execution history entry
type Entry struct {
	TaskFile     string        `json:"task_file"`
	ExecutedAt   time.Time     `json:"executed_at"`
	Success      bool          `json:"success"`
	Duration     time.Duration `json:"duration"`
	TaskCount    int           `json:"task_count"`
	ErrorMessage string        `json:"error_message,omitempty"`
	StartFrom    int           `json:"start_from,omitempty"`
	MaxRetries   int           `json:"max_retries,omitempty"`
	BranchName   string        `json:"branch_name,omitempty"`
}

// History manages task execution history
type History struct {
	Entries []Entry `json:"entries"`
}

// Load loads history from the history file
func Load(projectDir string) (*History, error) {
	historyPath := getHistoryPath(projectDir)

	// If history file doesn't exist, return empty history
	if _, err := os.Stat(historyPath); os.IsNotExist(err) {
		return &History{Entries: []Entry{}}, nil
	}

	data, err := os.ReadFile(historyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read history file: %w", err)
	}

	var history History
	if err := json.Unmarshal(data, &history); err != nil {
		return nil, fmt.Errorf("failed to parse history file: %w", err)
	}

	return &history, nil
}

// Save saves history to the history file
func (h *History) Save(projectDir string) error {
	historyPath := getHistoryPath(projectDir)

	// Create directory if it doesn't exist
	dir := filepath.Dir(historyPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create history directory: %w", err)
	}

	data, err := json.MarshalIndent(h, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal history: %w", err)
	}

	if err := os.WriteFile(historyPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write history file: %w", err)
	}

	return nil
}

// Add adds a new entry to the history
func (h *History) Add(entry Entry) {
	h.Entries = append(h.Entries, entry)
}

// GetLast returns the last N entries
func (h *History) GetLast(n int) []Entry {
	if n <= 0 {
		return []Entry{}
	}

	if n >= len(h.Entries) {
		return h.Entries
	}

	return h.Entries[len(h.Entries)-n:]
}

// GetFailed returns only failed entries
func (h *History) GetFailed() []Entry {
	var failed []Entry
	for _, entry := range h.Entries {
		if !entry.Success {
			failed = append(failed, entry)
		}
	}
	return failed
}

// GetSucceeded returns only successful entries
func (h *History) GetSucceeded() []Entry {
	var succeeded []Entry
	for _, entry := range h.Entries {
		if entry.Success {
			succeeded = append(succeeded, entry)
		}
	}
	return succeeded
}

// getHistoryPath returns the absolute path to the history file
func getHistoryPath(projectDir string) string {
	return filepath.Join(projectDir, historyDir, historyFile)
}

// Record is a convenience function to record a task execution
func Record(projectDir, taskFile, branchName string, success bool, duration time.Duration, taskCount, startFrom, maxRetries int, errorMsg string) error {
	history, err := Load(projectDir)
	if err != nil {
		return err
	}

	entry := Entry{
		TaskFile:     taskFile,
		ExecutedAt:   time.Now(),
		Success:      success,
		Duration:     duration,
		TaskCount:    taskCount,
		ErrorMessage: errorMsg,
		StartFrom:    startFrom,
		MaxRetries:   maxRetries,
		BranchName:   branchName,
	}

	history.Add(entry)

	return history.Save(projectDir)
}
