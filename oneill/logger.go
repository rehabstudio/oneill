package oneill

import (
	"fmt"
	"time"
)

func baseLogger(level string, message string) {
	fmt.Printf("%s [%s]: %s\n", time.Now().UTC().Format("2006-01-02T15:04:05.000Z"), level, message)
}

func LogDebug(message string) {
	if Config.LogLevel > 0 {
		return
	}
	baseLogger("DEBUG", message)
}

func LogInfo(message string) {
	if Config.LogLevel > 1 {
		return
	}
	baseLogger("INFO", message)
}

func LogWarning(message string) {
	if Config.LogLevel > 2 {
		return
	}
	baseLogger("WARNING", message)
}

func LogError(message string) {
	if Config.LogLevel > 3 {
		return
	}
	baseLogger("ERROR", message)
}
