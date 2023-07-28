package jsonlog

import (
	"io"
	"sync"
)

type Level int8

const (
	LevelInfo  Level = iota //Has the value of 0
	LeverError              //Has the value of 1 etc
	LevelFatal
	LevelOff
)

// Return a human-friendly string for the severity level
func (l Level) String() string {
	switch l {
	case LevelInfo:
		return "INFO"
	case LeverError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	default:
		return ""
	}
}

// Logger Define a custom logger type. This holds the destination that the log entries will  be written to.
type Logger struct {
	out      io.Writer
	minLevel Level
	mu       sync.Mutex
}

// New Return a new Logger instance which writes log entries
func New(out io.Writer, minLevel Level) *Logger {
	return &Logger{
		out:      out,
		minLevel: minLevel,
	}
}
