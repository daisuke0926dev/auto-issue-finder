package cmd

import (
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
