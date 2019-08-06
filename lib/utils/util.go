package utils

import (
	"fmt"
	"github.com/aybabtme/rgbterm"
	"os"
	"regexp"
	"strings"
)

// Check if file exists
func fileExists(path string) bool {
	info, _ := os.Stat(path)
	return info != nil && !info.IsDir()
}

// Labels object
type labels struct {
	addColon bool
	r        uint8
	g        uint8
	b        uint8
	colored  bool
}

func NewLabels(addColon bool, r, g, b uint8) *labels {
	lbl := new(labels)
	lbl.addColon = addColon
	lbl.r, lbl.g, lbl.b = r, g, b
	lbl.colored = true

	return lbl
}

func (lbl *labels) MapKeyToLabel(label string) string {
	unCamel, _ := regexp.Compile(`([A-Z]+)`)
	descore, _ := regexp.Compile(`[_]`)
	label = strings.Title(strings.ToLower(strings.TrimSpace(unCamel.ReplaceAllString(label, " $1"))))
	if lbl.colored {
		label = rgbterm.FgString(label, lbl.r, lbl.g, lbl.b)
	}
	label = descore.ReplaceAllString(label, " ")
	if lbl.addColon {
		label += ":"
	}
	return label
}

// Set coloring
func (lbl *labels) SetColored(isColored bool) {
	lbl.colored = isColored
}

// Console object
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
func (cns *console) CheckError(err error) {
	if err != nil {
		cns.ExitOnStderr(err.Error())
	}
}

func (cns *console) ExitOnStderr(message string) {
	os.Stderr.WriteString(fmt.Sprintf("Error: %s\n", message))
	os.Exit(1)
}

func (cns *console) ExitOnUnknown(message string) {
	cns.ExitOnStderr(fmt.Sprintf("%s %s", message, cns.hint))
}

// Console instance
var Console console

func init() {
	Console = *NewConsole()
}
