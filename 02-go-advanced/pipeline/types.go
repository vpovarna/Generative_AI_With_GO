package pipeline

import (
	"time"
)

type LogLevel string

const (
	INFO  = "INFO"
	ERROR = "ERROR"
	WARN  = "WARN"
	DEBUG = "DEBUG"
)

type LogEntry struct {
	Timestamp time.Time
	Level     LogLevel
	Message   string
	Source    string
}

// func GenerateLogs(ctx context.Context, sourceID string, rate time.Duration) <-chan LogEntry {
// }
