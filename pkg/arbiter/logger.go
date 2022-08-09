package arbiter

import "fmt"

// Logger is a simple system for logging errors
// and other such things
type Logger struct {
	Route string `json:"route"` //route defines where the logger writes to
}

// NewDefaultLogger returns a default Arbiter logger
func NewDefaultLogger() *Logger {
	return &Logger{}
}

func (l *Logger) Write(v ...any) error {
	if l.Route == "stdout" || l.Route == "std.Out" || l.Route == "" {
		fmt.Println(v...)
		return
	}
	// TODO: write v to log
	return nil
}

func (l *Logger) Writef(s string, v ...any) error {

	var towrite string

	if len(v) > 0 {
		towrite = fmt.Sprintf(s, v...)
	}
	// TODO: write string to log

	return nil
}
