package utils

import (
	"fmt"
	"log"
	"os"
)

const (
	_info = iota
	_error
	_warning
	_debug
	_fatal
)

type LoggerController struct {
	errors   bool
	warnings bool
	infos    bool
	debugs   bool
}

func NewLoggerController(nfo bool, warn bool, err bool, dbg bool) *LoggerController {
	controller := new(LoggerController)
	controller.errors = err
	controller.infos = nfo
	controller.warnings = warn
	controller.debugs = dbg

	return controller
}

// Put an information to the logger
func (logger *LoggerController) put(level int, message string, args ...string) {
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
	case _fatal:
		prefix = "FATAL"
	default:
		Console.ExitOnStderr(fmt.Sprintf("Unknown logging level: %d", level))
	}

	if len(args) > 0 {
		sargs := make([]interface{}, len(args))
		for idx := range sargs {
			sargs[idx] = args[idx]
		}
		message = fmt.Sprintf(message, sargs...)
	}

	log.Printf("[%s] %s\n", prefix, message)
	if level == _fatal {
		os.Exit(1)
	}
}

// Log info level
func (logger *LoggerController) Info(message string, args ...string) {
	if logger.infos {
		logger.put(_info, message, args...)
	}
}

// Log warning level
func (logger *LoggerController) Warning(message string, args ...string) {
	if logger.warnings {
		logger.put(_warning, message, args...)
	}
}

// Log error level
func (logger *LoggerController) Error(message string, args ...string) {
	if logger.errors {
		logger.put(_error, message, args...)
	}
}

// Log debug level
func (logger *LoggerController) Debug(message string, args ...string) {
	if logger.debugs {
		logger.put(_debug, message, args...)
	}
}

// Log fatal message and quit
func (logger *LoggerController) Fatal(message string, args ...string) {
	logger.put(_fatal, message, args...)
}
