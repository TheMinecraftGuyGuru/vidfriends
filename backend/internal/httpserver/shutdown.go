package httpserver

import "time"

// ShutdownTimeout controls how long to wait for graceful shutdowns.
var ShutdownTimeout = 10 * time.Second
