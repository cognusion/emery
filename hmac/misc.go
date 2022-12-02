package hmac

import (
	"io"
	"log"
)

// DebugOut is an optional logger to receive debugging messages
var DebugOut = log.New(io.Discard, "[DEBUG] ", 0)

// Error is an error type
type Error string

// Error returns the stringified version of Error
func (e Error) Error() string {
	return string(e)
}
