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

type console struct {
	hint string
}

// NewConsole constructor
func NewConsole() *console {
	cns := new(console)
	cns.hint = "Try --help for more details.\n"
	return cns
}

// Error checking
func (cns *console) checkError(err error) {
	if err != nil {
		cns.exitOnStderr(err.Error())
	}
}

func (cns *console) exitOnStderr(message string) {
	os.Stderr.WriteString(fmt.Sprintf("Error: %s\n", message))
	os.Exit(1)
}

func (cns *console) exitOnUnknown(message string) {
	cns.exitOnStderr(fmt.Sprintf("%s %s", message, cns.hint))
}

// Console instance
var Console console

func init() {
	Console = *NewConsole()
}
