package record

import (
	"errors"
	"fmt"
	"reflect"
)

// Record is a mutable envelope for record identity, data, and retrieval state.
type Record interface {
	Key() *Key
	Error() error
	Exists() bool
	SetError(error) Record
	Data() any
	HasChanged() bool
	MarkAsChanged()
}

type record struct {
	key     *Key
	err     error
	changed bool
	data    any
}

func (v *record) Exists() bool {
	if v.err != nil {
		if errors.Is(v.err, ErrNoError) {
			return true
		}
		if IsNotFound(v.err) {
			return false
		}
		panic(fmt.Errorf("an attempt to check if record exists for a record that has error: record.Key=%s; err: %s", v.Key(), v.err))
	}
	panic("an attempt to check if record exists before it was retrieved from database and SetError(error) called: record.Key=" + v.Key().String())
}

func (v *record) Key() *Key { return v.key }

func (v *record) HasChanged() bool { return v.changed }

func (v *record) MarkAsChanged() { v.changed = true }

func (v *record) Data() any {
	if v.err == nil {
		panic("an attempt to access record data before it was retrieved from database and SetError(error) called")
	}
	if errors.Is(v.err, ErrNoError) || IsNotFound(v.err) {
		return v.data
	}
	panic(fmt.Errorf("an attempt to retrieve data from a record with an error: %w", v.err))
}

func (v *record) Error() error {
	if v.err == nil || errors.Is(v.err, ErrNoError) || IsNotFound(v.err) {
		return nil
	}
	return v.err
}

func (v *record) SetError(err error) Record {
	if err == nil {
		v.err = ErrNoError
	} else {
		v.err = err
	}
	return v
}

// NewRecord creates an envelope with key and no data target.
func NewRecord(key *Key) Record { return newRecordWithOnlyKey(key) }

func newRecordWithOnlyKey(key *Key) *record {
	if key == nil {
		panic("parameter 'key' is required for record.NewRecord()")
	}
	if err := key.Validate(); err != nil {
		panic(fmt.Errorf("invalid key: %w", err))
	}
	return &record{key: key}
}

// NewRecordWithData creates an envelope carrying data the caller already has.
//
// The error is set to ErrNoError because the data was supplied, not retrieved:
// there is nothing to wait for, so Data() may be read immediately. This is the
// same reasoning NewRecordWithIncompleteKey already applies. Contrast NewRecord,
// which creates an empty envelope awaiting a load and deliberately leaves the
// error unset so Data() panics until a loader has said how the load went.
func NewRecordWithData(key *Key, data any) Record {
	record := newRecordWithOnlyKey(key)
	record.data = data
	record.err = ErrNoError
	return record
}

// NewRecordWithIncompleteKey creates an envelope for a record whose ID will be supplied later.
func NewRecordWithIncompleteKey(collection string, idKind reflect.Kind, data any) Record {
	return &record{key: NewIncompleteKey(collection, idKind, nil), data: data, err: ErrNoError}
}

// NewRecordWithoutKey creates an envelope for data that has no key.
func NewRecordWithoutKey(data any) Record {
	if data == nil {
		panic("data must not be nil")
	}
	t := reflect.TypeOf(data)
	switch t.Kind() {
	case reflect.Map:
		if t.Key().Kind() != reflect.String {
			panic("map key must be string")
		}
	case reflect.Slice:
	case reflect.Pointer:
		switch t.Elem().Kind() {
		case reflect.Map:
			panic("pointer to map is not allowed; pass the map by value")
		case reflect.Slice:
			panic("pointer to slice is not allowed; pass the slice by value")
		}
	default:
		panic("data must be a pointer, map[string]..., or slice")
	}
	// ErrNoError for the same reason as NewRecordWithData: the caller supplied
	// the data, so there is no retrieval to wait for.
	return &record{data: data, err: ErrNoError}
}

// AnyRecordWithError returns the first non-not-found record error.
func AnyRecordWithError(records ...Record) error {
	for _, record := range records {
		if err := record.Error(); err != nil && !IsNotFound(err) {
			return err
		}
	}
	return nil
}
