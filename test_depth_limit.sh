#!/bin/bash

echo "=== Testing Depth Limit in runCommand ==="

# Create a test script that simulates runCommand behavior
cat > /tmp/test_sleepship_depth.go << 'EOF'
package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

const maxRecursionDepth = 3

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

func testCommand(command string, depth int) {
	os.Setenv("SLEEPSHIP_DEPTH", fmt.Sprintf("%d", depth))

	isSleepshipCommand := strings.Contains(command, "sleepship") || strings.Contains(command, "./bin/sleepship")

	if isSleepshipCommand {
		currentDepth := getCurrentRecursionDepth()
		if currentDepth >= maxRecursionDepth {
			fmt.Printf("⚠️ Maximum recursion depth (%d) reached. Skipping sleepship command: %s\n", maxRecursionDepth, command)
			return
		}
		fmt.Printf("🔁 Executing recursive sleepship command (depth: %d -> %d)\n", currentDepth, currentDepth+1)
	}

	fmt.Printf("✅ Command would execute: %s (depth: %d)\n", command, getCurrentRecursionDepth())
}

func main() {
	commands := []string{
		"./bin/sleepship sync test.txt",
		"echo test",
		"sleepship sync another.txt",
	}

	for depth := 0; depth <= 4; depth++ {
		fmt.Printf("\n--- Testing at depth %d ---\n", depth)
		for _, cmd := range commands {
			testCommand(cmd, depth)
		}
	}
}
EOF

# Run the test
cd /tmp
go run test_sleepship_depth.go

echo -e "\n=== Test Complete ==="
