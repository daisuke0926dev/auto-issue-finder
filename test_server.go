// +build ignore

package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func main() {
	// Test root endpoint
	fmt.Println("Testing / endpoint...")
	resp1, err := http.Get("http://localhost:8080/")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer resp1.Body.Close()
	body1, _ := io.ReadAll(resp1.Body)
	fmt.Printf("Status: %d\n", resp1.StatusCode)
	fmt.Printf("Response: %s\n\n", string(body1))

	// Test health endpoint
	fmt.Println("Testing /health endpoint...")
	resp2, err := http.Get("http://localhost:8080/health")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer resp2.Body.Close()
	body2, _ := io.ReadAll(resp2.Body)
	fmt.Printf("Status: %d\n", resp2.StatusCode)
	fmt.Printf("Content-Type: %s\n", resp2.Header.Get("Content-Type"))
	fmt.Printf("Response: %s\n", string(body2))
}
