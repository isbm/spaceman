package main

import (
	"os"
)

// Check if file exists
func fileExists(path string) bool {
	info, _ := os.Stat(path)
	return info != nil && !info.IsDir()
}
