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

// ErrRecordExists indicates that an insert failed because a record with the
// same key already exists. It is the mirror image of ErrRecordNotFound: an
// adapter that can reliably tell "this key is taken" from any other insert
// failure wraps this sentinel (e.g. with fmt.Errorf("%w: %s", ErrRecordExists,
// key)) so callers can distinguish the two with errors.Is or IsAlreadyExists,
// instead of having to treat every insert failure as a duplicate key.
var ErrRecordExists = errors.New("record already exists")

// IsAlreadyExists reports whether err represents a record that already
// exists. It positively identifies a duplicate-key conflict: true means the
// adapter itself classified this failure as one. False is NOT proof the
// failure was something else — an adapter that cannot yet distinguish a
// duplicate key from any other insert failure never returns ErrRecordExists,
// so it will report false here even on an actual conflict. Callers should use
// this predicate to detect a confirmed conflict, and keep treating an
// unclassified insert failure conservatively (i.e. not assume "not a
// conflict") until every adapter they rely on has adopted the sentinel.
func IsAlreadyExists(err error) bool {
	return err != nil && errors.Is(err, ErrRecordExists)
}
