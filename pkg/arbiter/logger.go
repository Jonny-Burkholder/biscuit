package arbiter

import (
	"fmt"
	"io"
	"log"
	"os"
)

// Logger is a simple system for logging errors
// and other such things
type Logger struct {
	Logger log.Logger
}

// NewDefaultLogger returns a default Arbiter logger
func NewDefaultLogger() *Logger {
	return &Logger{
		log.New(os.Stderr, "", nil),
	}
}

// NewLogger, yeah
func NewLogger(w io.Writer) *Logger {
	return &Logger{
		log.New(w, "", nil),
	}
}

func (l *Logger) Write(format string, v ...any) {
	if l.Route == "stdout" || l.Route == "std.Out" || l.Route == "" {
		fmt.Println(v...)
		return
	}
	// TODO: write v to log
	return nil
}

func (l *Logger) Trace(format string, v ...any) {}

func (l *Logger) Debug(format string, v ...any) {}

func (l *Logger) Info(format string, v ...any) {}

func (l *Logger) Warn(format string, v ...any) {}

func (l *Logger) Fatal(format string, v ...any) {}

func (l *Logger) Panic(format string, v ...any) {}
