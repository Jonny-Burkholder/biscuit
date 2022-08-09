package router

import "net/http"

// package router implements a custom multiplexer designed to
// be used in conjunction with biscuit and arbiter

// Logger interface allows us to use a default Arbiter logger,
// or a custom one
type Logger interface {
	write(v ...any)
}

// Router keeps track of routes and their various methods
type Router struct {
	DefaultError http.Handler // this is a configureable route for handling 404, etc

	Logger *Logger

	Routes map[string][]http.Handler
}
