package logger

import (
	"fmt"
	"time"
)

type StdOutLogger struct {
	logLevel int
}

func NewStdOutLogger(logLevel int) *StdOutLogger {
	stdOutLogger := StdOutLogger{logLevel: logLevel}
	return &stdOutLogger
}

// baseLogger handles formatting and output for all other loggers
func (s *StdOutLogger) baseLogger(threshold int, level string, message string) {
	if s.logLevel > threshold {
		return
	}
	timestamp := time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
	fmt.Printf("%s [%s]: %s\n", timestamp, level, message)
}

func (s *StdOutLogger) Debug(message string) error {
	s.baseLogger(1, "DEBUG", message)
	return nil
}

func (s *StdOutLogger) Info(message string) error {
	s.baseLogger(2, "INFO", message)
	return nil
}

func (s *StdOutLogger) Warning(message string) error {
	s.baseLogger(3, "WARNING", message)
	return nil
}

func (s *StdOutLogger) Error(message string) error {
	s.baseLogger(4, "ERROR", message)
	return nil
}
