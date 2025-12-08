package history

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadAndSave(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "sleepship-history-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Load empty history
	hist, err := Load(tempDir)
	if err != nil {
		t.Fatalf("Failed to load history: %v", err)
	}

	if len(hist.Entries) != 0 {
		t.Errorf("Expected 0 entries, got %d", len(hist.Entries))
	}

	// Add entry
	entry := Entry{
		TaskFile:   "tasks-test.txt",
		ExecutedAt: time.Now(),
		Success:    true,
		Duration:   5 * time.Second,
		TaskCount:  3,
		StartFrom:  1,
		MaxRetries: 3,
		BranchName: "feature/test",
	}
	hist.Add(entry)

	// Save history
	if err := hist.Save(tempDir); err != nil {
		t.Fatalf("Failed to save history: %v", err)
	}

	// Check if file exists
	historyPath := filepath.Join(tempDir, historyDir, historyFile)
	if _, err := os.Stat(historyPath); os.IsNotExist(err) {
		t.Errorf("History file was not created: %s", historyPath)
	}

	// Load history again
	hist2, err := Load(tempDir)
	if err != nil {
		t.Fatalf("Failed to load history again: %v", err)
	}

	if len(hist2.Entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(hist2.Entries))
	}

	// Check entry values
	loaded := hist2.Entries[0]
	if loaded.TaskFile != entry.TaskFile {
		t.Errorf("TaskFile mismatch: expected %s, got %s", entry.TaskFile, loaded.TaskFile)
	}
	if loaded.Success != entry.Success {
		t.Errorf("Success mismatch: expected %v, got %v", entry.Success, loaded.Success)
	}
	if loaded.TaskCount != entry.TaskCount {
		t.Errorf("TaskCount mismatch: expected %d, got %d", entry.TaskCount, loaded.TaskCount)
	}
}

func TestGetLast(t *testing.T) {
	hist := &History{
		Entries: []Entry{
			{TaskFile: "task1.txt", Success: true},
			{TaskFile: "task2.txt", Success: true},
			{TaskFile: "task3.txt", Success: false},
			{TaskFile: "task4.txt", Success: true},
			{TaskFile: "task5.txt", Success: true},
		},
	}

	// Get last 3
	last3 := hist.GetLast(3)
	if len(last3) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(last3))
	}
	if last3[0].TaskFile != "task3.txt" {
		t.Errorf("Expected task3.txt, got %s", last3[0].TaskFile)
	}
	if last3[2].TaskFile != "task5.txt" {
		t.Errorf("Expected task5.txt, got %s", last3[2].TaskFile)
	}

	// Get last 0
	last0 := hist.GetLast(0)
	if len(last0) != 0 {
		t.Errorf("Expected 0 entries, got %d", len(last0))
	}

	// Get last 10 (more than available)
	last10 := hist.GetLast(10)
	if len(last10) != 5 {
		t.Errorf("Expected 5 entries, got %d", len(last10))
	}
}

func TestGetFailed(t *testing.T) {
	hist := &History{
		Entries: []Entry{
			{TaskFile: "task1.txt", Success: true},
			{TaskFile: "task2.txt", Success: false, ErrorMessage: "error1"},
			{TaskFile: "task3.txt", Success: false, ErrorMessage: "error2"},
			{TaskFile: "task4.txt", Success: true},
			{TaskFile: "task5.txt", Success: false, ErrorMessage: "error3"},
		},
	}

	failed := hist.GetFailed()
	if len(failed) != 3 {
		t.Errorf("Expected 3 failed entries, got %d", len(failed))
	}

	for _, entry := range failed {
		if entry.Success {
			t.Errorf("Expected failed entry, got successful: %s", entry.TaskFile)
		}
	}
}

func TestGetSucceeded(t *testing.T) {
	hist := &History{
		Entries: []Entry{
			{TaskFile: "task1.txt", Success: true},
			{TaskFile: "task2.txt", Success: false, ErrorMessage: "error1"},
			{TaskFile: "task3.txt", Success: true},
			{TaskFile: "task4.txt", Success: true},
			{TaskFile: "task5.txt", Success: false, ErrorMessage: "error2"},
		},
	}

	succeeded := hist.GetSucceeded()
	if len(succeeded) != 3 {
		t.Errorf("Expected 3 successful entries, got %d", len(succeeded))
	}

	for _, entry := range succeeded {
		if !entry.Success {
			t.Errorf("Expected successful entry, got failed: %s", entry.TaskFile)
		}
	}
}

func TestRecord(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "sleepship-history-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Record entry
	err = Record(tempDir, "tasks-test.txt", "feature/test", true, 10*time.Second, 5, 1, 3, "")
	if err != nil {
		t.Fatalf("Failed to record: %v", err)
	}

	// Load and verify
	hist, err := Load(tempDir)
	if err != nil {
		t.Fatalf("Failed to load history: %v", err)
	}

	if len(hist.Entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(hist.Entries))
	}

	entry := hist.Entries[0]
	if entry.TaskFile != "tasks-test.txt" {
		t.Errorf("TaskFile mismatch: expected tasks-test.txt, got %s", entry.TaskFile)
	}
	if !entry.Success {
		t.Errorf("Expected success=true, got false")
	}
	if entry.TaskCount != 5 {
		t.Errorf("TaskCount mismatch: expected 5, got %d", entry.TaskCount)
	}

	// Record another entry
	err = Record(tempDir, "tasks-test2.txt", "feature/test2", false, 5*time.Second, 3, 1, 3, "test error")
	if err != nil {
		t.Fatalf("Failed to record second entry: %v", err)
	}

	// Load and verify
	hist, err = Load(tempDir)
	if err != nil {
		t.Fatalf("Failed to load history: %v", err)
	}

	if len(hist.Entries) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(hist.Entries))
	}

	entry2 := hist.Entries[1]
	if entry2.TaskFile != "tasks-test2.txt" {
		t.Errorf("TaskFile mismatch: expected tasks-test2.txt, got %s", entry2.TaskFile)
	}
	if entry2.Success {
		t.Errorf("Expected success=false, got true")
	}
	if entry2.ErrorMessage != "test error" {
		t.Errorf("ErrorMessage mismatch: expected 'test error', got '%s'", entry2.ErrorMessage)
	}
}
