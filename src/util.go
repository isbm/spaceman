package main

import (
	"fmt"
	"os"
)

// Check if file exists
func fileExists(path string) bool {
	info, _ := os.Stat(path)
	return info != nil && !info.IsDir()
}

// Standard finaliser
func endWithHint(message string) {
	os.Stderr.WriteString(fmt.Sprintf("%s Try --help for more details.\n", message))
	os.Exit(1)
}
