package injecttag

import (
	"log"
)

// Verbose controls whether verbose logging is enabled
var Verbose = false

// Logf logs a formatted message if verbose mode is enabled
func Logf(format string, v ...interface{}) {
	if !Verbose {
		return
	}
	log.Printf(format, v...)
}

// Internal lowercase versions for backward compatibility within package
var verbose = &Verbose

func logf(format string, v ...interface{}) {
	Logf(format, v...)
}
