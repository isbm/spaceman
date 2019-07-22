package main

import (
	"fmt"
	"log"
)

const (
	_info = iota
	_error
	_warning
	_debug
)

type loggerController struct {
	errors bool
	infos  bool
	debugs bool
}

func LoggerController(errors bool, infos bool, debugs bool) *loggerController {
	controller := new(loggerController)
	controller.errors = errors
	controller.infos = infos
	controller.debugs = debugs

	return controller
}

// Put an information to the logger
func (logger *loggerController) put(level int, message string) {
	var prefix string
	switch level {
	case _info:
		prefix = "INFO"
	case _warning:
		prefix = "WARNING"
	case _error:
		prefix = "ERROR"
	case _debug:
		prefix = "DEBUG"
	default:
		Console.exitOnStderr(fmt.Sprintf("Unknown logging level: %d", level))
	}

	log.Printf("[%s] %s\n", prefix, message)
}

// Log info level
func (logger *loggerController) Info(message string) {
	logger.put(_info, message)
}

// Log warning level
func (logger *loggerController) Warning(message string) {
	logger.put(_warning, message)
}

// Log error level
func (logger *loggerController) Error(message string) {
	logger.put(_error, message)
}

// Log debug level
func (logger *loggerController) Debug(message string) {
	logger.put(_debug, message)
}
