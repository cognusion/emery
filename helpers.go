package main

import (
	"io"
	"log"
	"os"
)

var (
	// OutFormat is a log.Logger format used by default
	OutFormat = log.Ldate | log.Ltime | log.Lshortfile
	// DebugOut is a log.Logger for debug messages
	DebugOut = log.New(io.Discard, "[DEBUG] ", 0)
	// TimingOut is a log.Logger for timing-related debug messages. DEPRECATED
	TimingOut = log.New(io.Discard, "[TIMING] ", 0)
	// ErrorOut is a log.Logger for error messages
	ErrorOut = log.New(os.Stderr, "", OutFormat)
	// AccessOut is a log.Logger for access logging. PLEASE DO NOT USE THIS DIRECTLY
	AccessOut = log.New(os.Stdout, "", 0)
	// CommonOut is a log.Logger for Apache "common log format" logging. PLEASE DO NOT USE THIS DIRECTLY
	CommonOut = log.New(io.Discard, "", 0)
	// SlowOut is a log.Logger for slow request information
	SlowOut = log.New(io.Discard, "", 0)
)

// Error is an error type
type Error string

// Error returns the stringified version of Error
func (e Error) Error() string {
	return string(e)
}
