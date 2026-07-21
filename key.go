package record

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// Key represents the full identity path of a record.
type Key struct {
	parent     *Key
	collection string
	ID         any
	IDKind     reflect.Kind
}

// KeyOption configures a Key during construction.
type KeyOption func(*Key) error

var idCharsReplacer = strings.NewReplacer(
	".", "%2E",
	"$", "%24",
	"#", "%23",
	"[", "%5B",
	"]", "%5D",
	"/", "%2F",
)

// EscapeID escapes a record ID for use in a key path.
func EscapeID(id string) string {
	return idCharsReplacer.Replace(id)
}

// String returns the serialized path of k.
func (k *Key) String() string {
	key := k
	if err := key.Validate(); err != nil {
		panic(fmt.Sprintf("will not generate path for invalid key: %v", err))
	}
	s := make([]string, 0, key.Level()*2)
	for {
		id := EscapeID(fmt.Sprintf("%v", key.ID))
		s = append(s, id, key.collection)
		if key.parent == nil {
			break
		}
		key = key.parent
	}
	return reverseStringsJoin(s, "/")
}

// CollectionPath returns the collection path of k.
func (k *Key) CollectionPath() string {
	key := k
	var s []string
	for {
		if strings.TrimSpace(key.collection) == "" {
			panic("k is referencing an empty recordsetSource")
		}
		s = append(s, key.collection)
		if key.parent == nil {
			break
		}
		key = key.parent
	}
	return reverseStringsJoin(s, "/")
}

func reverseStringsJoin(elems []string, sep string) string {
	if len(elems) == 0 {
		return ""
	}
	n := len(sep) * (len(elems) - 1)
	for _, elem := range elems {
		n += len(elem)
	}
	var b strings.Builder
	b.Grow(n)
	for i := len(elems) - 1; i >= 0; i-- {
		if _, err := b.WriteString(elems[i]); err != nil {
			panic(err)
		}
		if i > 0 {
			if _, err := b.WriteString(sep); err != nil {
				panic(err)
			}
		}
	}
	return b.String()
}

// Level returns the number of ancestors of k.
func (k *Key) Level() int {
	if k.parent == nil {
		return 0
	}
	return k.parent.Level() + 1
}

// Parent returns the parent key, if any.
func (k *Key) Parent() *Key {
	return k.parent
}

// Collection returns k's collection name.
func (k *Key) Collection() string {
	return k.collection
}

// Validate validates k and its ancestors.
func (k *Key) Validate() error {
	if strings.TrimSpace(k.collection) == "" {
		return errors.New("key must have `recordsetSource` field value")
	}
	if k.parent != nil {
		return k.parent.Validate()
	}
	if fields, ok := k.ID.([]FieldVal); ok {
		for i, field := range fields {
			if err := field.Validate(); err != nil {
				return fmt.Errorf("key has a invalid referencing to a field value #%v: %w", i, err)
			}
		}
	}
	if id, ok := k.ID.(interface{ Validate() error }); ok {
		return id.Validate()
	}
	return nil
}

// NewKeyWithParentAndID creates a key below parent with id.
func NewKeyWithParentAndID[T comparable](parent *Key, collection string, id T) *Key {
	key := NewKeyWithID(collection, id)
	key.parent = parent
	return key
}

// NewKeyWithID creates a key with id.
func NewKeyWithID[T comparable](collection string, id T) *Key {
	if collection == "" {
		panic("recordsetSource is a required parameter")
	}
	return &Key{collection: collection, ID: id}
}

// NewIncompleteKey creates a key whose ID will be supplied later.
func NewIncompleteKey(collection string, idKind reflect.Kind, parent *Key) *Key {
	if idKind == reflect.Invalid {
		panic("idKind == reflect.Invalid")
	}
	return &Key{parent: parent, collection: collection, IDKind: idKind}
}

// WithKeyID sets a key ID during construction.
func WithKeyID[T comparable](id T) KeyOption {
	return func(key *Key) error {
		key.ID = id
		return nil
	}
}

// WithFields sets composite key fields during construction.
func WithFields(fields []FieldVal) KeyOption {
	return func(key *Key) error {
		key.ID = fields
		return nil
	}
}

// WithParentKey sets a parent key during construction.
func WithParentKey(parent *Key) KeyOption {
	if parent == nil {
		panic("parent == nil")
	}
	return func(key *Key) error {
		key.parent = parent
		return nil
	}
}

// WithStringID sets a string key ID during construction.
func WithStringID(id string) KeyOption { return WithKeyID(id) }

// WithIntID sets an integer key ID during construction.
func WithIntID(id int) KeyOption { return WithKeyID(id) }

// NewKeyWithFields creates a key with a composite ID.
func NewKeyWithFields(collection string, fields ...FieldVal) *Key {
	return &Key{collection: collection, ID: fields}
}

// NewKeyWithOptions creates a key configured by options.
func NewKeyWithOptions(collection string, options ...KeyOption) (*Key, error) {
	if collection == "" {
		return nil, errors.New("recordsetSource is a required parameter")
	}
	key := &Key{collection: collection}
	if err := setKeyOptions(key, options...); err != nil {
		return nil, err
	}
	return key, nil
}

func setKeyOptions(key *Key, options ...KeyOption) error {
	for _, option := range options {
		if err := option(key); err != nil {
			return err
		}
	}
	return nil
}

// EqualKeys reports whether two key paths are equal.
func EqualKeys(k1, k2 *Key) bool {
	if k1 == nil && k2 == nil {
		return true
	}
	if k1 == nil || k2 == nil {
		return false
	}
	k1s := make([]*Key, 0, k1.Level())
	k2s := make([]*Key, 0, k2.Level())
	panicIfCircular := func(key *Key, keys []*Key) {
		for _, prior := range keys {
			if prior.ID == key.ID && prior.collection == key.collection {
				panic(fmt.Sprintf("circular key: %s=%v", prior.collection, prior.ID))
			}
		}
	}
	for {
		if k1 == nil && k2 == nil {
			return true
		}
		if k1 == nil || k2 == nil || k1.Collection() != k2.Collection() || k1.ID != k2.ID {
			return false
		}
		k1s = append(k1s, k1)
		k2s = append(k2s, k2)
		if k1 = k1.Parent(); k1 != nil {
			panicIfCircular(k1, k1s)
		}
		if k2 = k2.Parent(); k2 != nil {
			panicIfCircular(k2, k2s)
		}
	}
}

// Equal reports whether k and other are equal.
func (k *Key) Equal(other *Key) bool { return EqualKeys(k, other) }
