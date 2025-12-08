package main

import (
	"fmt"
	"os"
)

// Simplified Task struct for testing
type Task struct {
	Title       string
	Description string
	Command     string
}

func main() {
	// Simulate parsed tasks
	tasks := []Task{
		{Title: "Task 1: Test task 1", Description: "Description 1"},
		{Title: "Task 2: Test task 2", Description: "Description 2"},
		{Title: "Task 3: Test task 3", Description: "Description 3"},
		{Title: "Task 4: Test task 4", Description: "Description 4"},
		{Title: "Task 5: Test task 5", Description: "Description 5"},
	}

	// Test different startFrom values
	testStartFrom := []int{1, 3, 5}

	for _, startFrom := range testStartFrom {
		fmt.Printf("\n=== Testing with startFrom=%d ===\n", startFrom)
		fmt.Printf("📋 Total tasks: %d\n", len(tasks))
		if startFrom > 1 {
			fmt.Printf("⏩ Starting from task: %d\n", startFrom)
		}
		fmt.Println()

		for i, task := range tasks {
			taskNum := i + 1

			// Skip tasks before startFrom
			if taskNum < startFrom {
				fmt.Printf("⏭️  Skipping task %d/%d (start-from=%d): %s\n", taskNum, len(tasks), startFrom, task.Title)
				continue
			}

			fmt.Printf("▶️  Executing task %d/%d: %s\n", taskNum, len(tasks), task.Title)
		}
	}

	fmt.Println("\n✅ Skip logic test completed successfully!")
	os.Exit(0)
}
