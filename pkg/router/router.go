package router

import (
	"net/http"

	"github.com/Jonny-Burkholder/biscuit/pkg/arbiter"
)

// package router implements a custom multiplexer designed to
// be used in conjunction with biscuit and arbiter

// Logger interface allows us to use a default Arbiter logger,
// or a custom one
type Logger interface {
	Debug(format string, v ...any)
	Info(format string, v ...any)
	Fatal(format string, v ...any)
}

// Router keeps track of routes and their various methods
type Router struct {
	DefaultError http.Handler // this is a configureable route for handling 404, etc

	Logger *Logger

	Routes map[string]map[string]http.Handler

	// Groups map[string]*Group
}

/*
// Group will eventually be cool and do cool things
type Group struct {
	BaseRoute string
	Routes map[string]map[string]http.Handler // basically this goes Routes[method][route]http.Handler
	Use []interface{} // TODO: make this not suck. Basically this is middleware for the group to use
}
*/

// NewRouter takes a pointer to a logger as an argument, and returns a pointer
// to a new router. If the logger pointer is nil, the router is created with
// a default arbiter logger
func NewRouter(logger *Logger) *Router {
	if logger == nil {
		logger = arbiter.NewDefaultLogger()
	}
	// groups := make(map[string]*Group)
	// groups["default"] = defaultGroup()
	return &Router{
		Logger: logger,
		Routes: make(map[string]map[string]http.Handler),
		// Groups: groups
	}
}

func (r *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO: implement this
}

/*
// Group creates a new group based on the baseroute
func (r *Router) Group(baseroute string) {
	r.Groups[baseroute] = newGroup(baseroute)
}

// newGroup does what it says it does
func newGroup(baseroute string) *Group {
	return &Group{
		BaseRoute: baseroute,
		Routes: make(map[string]map[string]http.Handler,
		Use: make([]interface{})
	}
}
*/

// GET() adds a handler with method GET to the router
func (r *Router) GET(route string, handler http.Handler) {
	r.Routes["get"][route] = handler
}

// POST() adds a handler with method POST to the router
func (r *Router) POST(route string, handler http.Handler) {
	r.Routes["post"][route] = handler
}

// PUT() adds a handler with method PUT to the router
func (r *Router) PUT(route string, handler http.Handler) {
	r.Routes["put"][route] = handler
}

// DELETE() adds a handler with method DELETE to the router
func (r *Router) DELETE(route string, handler http.Handler) {
	r.Routes["delete"][route] = handler
}
