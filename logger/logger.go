package logger

import (
	"fmt"
	"time"
)

// configured log level for the application
var logLevel int = 0

// correctly initialise logging level from config
func InitLogger(level int) {
	logLevel = level
}

// baseLogger handles formatting and output for all other loggers
func baseLogger(level string, message string) {
	fmt.Printf("%s [%s]: %s\n", time.Now().UTC().Format("2006-01-02T15:04:05.000Z"), level, message)
}

func LogDebug(message string) {
	if logLevel > 0 {
		return
	}
	baseLogger("DEBUG", message)
}

func LogInfo(message string) {
	if logLevel > 1 {
		return
	}
	baseLogger("INFO", message)
}

func LogWarning(message string) {
	if logLevel > 2 {
		return
	}
	baseLogger("WARNING", message)
}

func LogError(message string) {
	if logLevel > 3 {
		return
	}
	baseLogger("ERROR", message)
}

func LogFatal(message string) {
	if logLevel > 4 {
		return
	}
	baseLogger("FATAL", message)
}
