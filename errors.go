package record

import (
	"errors"
)

// ErrNoError is the internal success sentinel used by Record state.
var ErrNoError = errors.New("no error")

// ErrRecordNotFound indicates that a requested record does not exist.
var ErrRecordNotFound = errors.New("record not found")

// IsNotFound reports whether err represents a missing record.
func IsNotFound(err error) bool {
	return err != nil && errors.Is(err, ErrRecordNotFound)
}
