package record

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestKeyPathAndParent(t *testing.T) {
	parent := NewKeyWithID("tenants", "acme/uk")
	key := NewKeyWithParentAndID(parent, "users", 42)
	if got, want := key.String(), "tenants/acme%2Fuk/users/42"; got != want {
		t.Fatalf("Key.String() = %q, want %q", got, want)
	}
	if key.Parent() != parent || key.Level() != 1 || key.Collection() != "users" {
		t.Fatal("key hierarchy was not preserved")
	}
}

func TestNewKeyWithOptions(t *testing.T) {
	parent := NewKeyWithID("tenants", "acme")
	key, err := NewKeyWithOptions("users", WithKeyID("u1"), WithParentKey(parent))
	if err != nil {
		t.Fatal(err)
	}
	if got, want := key.String(), "tenants/acme/users/u1"; got != want {
		t.Fatalf("Key.String() = %q, want %q", got, want)
	}

	incomplete := NewIncompleteKey("users", reflect.String, nil)
	if incomplete.ID != nil || incomplete.IDKind != reflect.String {
		t.Fatal("incomplete key did not retain its expected ID kind")
	}
}

// TestKeyConstructorsKeepIncompleteKeys verifies
// AC:key-constructors-keep-incomplete-keys.
func TestKeyConstructorsKeepIncompleteKeys(t *testing.T) {
	k := NewKeyWithID("users", "")
	if got, want := k.String(), "users/"; got != want {
		t.Fatalf("Key.String() = %q, want %q", got, want)
	}
	if err := k.Validate(); err != nil {
		t.Fatalf("Key.Validate() = %v, want nil", err)
	}

	parent := NewKeyWithID("spaces", "s1")
	child := NewKeyWithParentAndID(parent, "ext", "")
	if got, want := child.String(), "spaces/s1/ext/"; got != want {
		t.Fatalf("Key.String() = %q, want %q", got, want)
	}

	incomplete := NewIncompleteKey("users", reflect.String, nil)
	if got, want := incomplete.String(), "users/"; got != want {
		t.Fatalf("NewIncompleteKey(...).String() = %q, want %q", got, want)
	}
	if err := incomplete.Validate(); err != nil {
		t.Fatalf("NewIncompleteKey(...).Validate() = %v, want nil", err)
	}
}

// TestNewKeyWithIDPanicsOnInvalidStringID verifies the panicking half of
// AC:key-constructors-keep-incomplete-keys, and that the panic value is an
// error a recovering caller can inspect with errors.Is (not just a string).
func TestNewKeyWithIDPanicsOnInvalidStringID(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("NewKeyWithID did not panic on a string id containing '%'")
		}
		err, ok := r.(error)
		if !ok {
			t.Fatalf("panic value = %#v, want an error", r)
		}
		if !errors.Is(err, ErrInvalidStringID) {
			t.Fatalf("panic error = %v, want ErrInvalidStringID", err)
		}
		if !strings.Contains(err.Error(), ErrInvalidStringID.Error()) {
			t.Fatalf("panic message = %q, want it to name %v", err.Error(), ErrInvalidStringID)
		}
	}()
	NewKeyWithID("users", "a%2Fb")
}

// TestNewKeyWithParentAndIDPanicsOnInvalidStringID verifies that
// NewKeyWithParentAndID inherits NewKeyWithID's id validation.
func TestNewKeyWithParentAndIDPanicsOnInvalidStringID(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("NewKeyWithParentAndID did not panic on a string id containing '%'")
		}
		if err, ok := r.(error); !ok || !errors.Is(err, ErrInvalidStringID) {
			t.Fatalf("panic value = %#v, want an error satisfying errors.Is(_, ErrInvalidStringID)", r)
		}
	}()
	NewKeyWithParentAndID(NewKeyWithID("tenants", "t1"), "users", "a%2Fb")
}

// namedStringID is a named type whose underlying type is string, used to
// verify that id validation is not bypassed by a `id.(string)` type
// assertion that only matches the literal `string` type.
type namedStringID string

// TestNamedStringTypeIDIsValidated verifies that a named string type gets
// the same '%' validation as a plain string, across every id-accepting API.
func TestNamedStringTypeIDIsValidated(t *testing.T) {
	t.Run("NewKeyWithID panics", func(t *testing.T) {
		defer func() {
			r := recover()
			if r == nil {
				t.Fatal("NewKeyWithID did not panic on a named-string-type id containing '%'")
			}
			if err, ok := r.(error); !ok || !errors.Is(err, ErrInvalidStringID) {
				t.Fatalf("panic value = %#v, want an error satisfying errors.Is(_, ErrInvalidStringID)", r)
			}
		}()
		NewKeyWithID("users", namedStringID("a%2Fb"))
	})

	t.Run("WithKeyID returns an error", func(t *testing.T) {
		_, err := NewKeyWithOptions("users", WithKeyID(namedStringID("a%b")))
		if !errors.Is(err, ErrInvalidStringID) {
			t.Fatalf("NewKeyWithOptions(...) error = %v, want ErrInvalidStringID", err)
		}
	})

	t.Run("Validate rejects it", func(t *testing.T) {
		k := &Key{collection: "users", ID: namedStringID("a%b")}
		if err := k.Validate(); !errors.Is(err, ErrInvalidStringID) {
			t.Fatalf("Validate() = %v, want ErrInvalidStringID", err)
		}
	})

	t.Run("empty named-string id stays a legal incomplete key", func(t *testing.T) {
		k := NewKeyWithID("users", namedStringID(""))
		if err := k.Validate(); err != nil {
			t.Fatalf("Validate() = %v, want nil", err)
		}
	})
}

// TestNewKeyWithOptionsReturnsErrInvalidStringID verifies the non-panicking
// half of AC:key-constructors-keep-incomplete-keys for the WithKeyID path.
func TestNewKeyWithOptionsReturnsErrInvalidStringID(t *testing.T) {
	_, err := NewKeyWithOptions("users", WithKeyID("a%b"))
	if !errors.Is(err, ErrInvalidStringID) {
		t.Fatalf("NewKeyWithOptions(...) error = %v, want ErrInvalidStringID", err)
	}
}

// TestKeyStringNeverPanicsOnInvalidID verifies that String() does not
// validate its receiver, even when a key was assembled (bypassing the
// validating constructors, as only in-package code can) with an id that
// would fail Validate.
func TestKeyStringNeverPanicsOnInvalidID(t *testing.T) {
	k := &Key{collection: "users", ID: "a%b"}
	got := k.String()
	if got == "" {
		t.Fatal("Key.String() returned empty string unexpectedly")
	}
}

// TestNilKeyStringDoesNotPanic verifies that a nil *Key's String() method
// (a receiver value a caller can produce, e.g. from a zero-value field or a
// failed lookup) returns without panicking.
func TestNilKeyStringDoesNotPanic(t *testing.T) {
	var k *Key
	if got, want := k.String(), ""; got != want {
		t.Fatalf("(*Key)(nil).String() = %q, want %q", got, want)
	}
}

// TestTypedNilIDPrintsAsEmptySegment verifies that an id holding a typed nil
// (a non-nil `any` whose dynamic value is nil, e.g. a nil pointer) formats
// the same as a nil or empty id, rather than printing "<nil>".
func TestTypedNilIDPrintsAsEmptySegment(t *testing.T) {
	var nilPtr *int
	k := &Key{collection: "users", ID: nilPtr}
	if got, want := k.String(), "users/"; got != want {
		t.Fatalf("Key.String() = %q, want %q", got, want)
	}
}

// TestKeyValidateChecksEveryLevel verifies that Validate applies the '%'
// check to a non-empty string id at every level of the key, not just the
// root, and still reports a missing collection name.
func TestKeyValidateChecksEveryLevel(t *testing.T) {
	invalidChild := &Key{collection: "users", ID: "a%b"}
	if err := invalidChild.Validate(); !errors.Is(err, ErrInvalidStringID) {
		t.Fatalf("Validate() = %v, want ErrInvalidStringID", err)
	}

	invalidParent := &Key{collection: "spaces", ID: "s1%2"}
	child := NewKeyWithParentAndID(NewKeyWithID("tenants", "t1"), "users", "u1")
	child.parent = invalidParent
	if err := child.Validate(); !errors.Is(err, ErrInvalidStringID) {
		t.Fatalf("Validate() on child with invalid ancestor = %v, want ErrInvalidStringID", err)
	}

	missingCollection := &Key{ID: "x"}
	if err := missingCollection.Validate(); err == nil {
		t.Fatal("Validate() on key with empty collection = nil, want an error")
	}
}

// TestKeyValidateChecksFieldValAndCustomValidator verifies the composite-key
// and custom-id branches of Validate.
func TestKeyValidateChecksFieldValAndCustomValidator(t *testing.T) {
	badFields := NewKeyWithFields("countries", FieldVal{Name: "", Value: "us"})
	if err := badFields.Validate(); err == nil {
		t.Fatal("Validate() with an unnamed field value = nil, want an error")
	}

	goodFields := NewKeyWithFields("countries", FieldVal{Name: "country", Value: "us"})
	if err := goodFields.Validate(); err != nil {
		t.Fatalf("Validate() with a valid field value = %v, want nil", err)
	}

	key := &Key{collection: "users", ID: validatingID{err: errors.New("boom")}}
	if err := key.Validate(); err == nil || err.Error() != "boom" {
		t.Fatalf("Validate() with a failing custom id = %v, want %q", err, "boom")
	}
}

type validatingID struct{ err error }

func (v validatingID) Validate() error { return v.err }

// TestNewKeyWithIDPanicsOnEmptyCollection documents the pre-existing
// empty-collection guard NewKeyWithID's id validation sits next to.
func TestNewKeyWithIDPanicsOnEmptyCollection(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("NewKeyWithID did not panic on an empty collection")
		}
	}()
	NewKeyWithID("", "id")
}

func TestValidateStringID(t *testing.T) {
	for _, id := range []string{"competition-1", "tenant/user", "space.with$markers#[1]"} {
		if err := ValidateStringID(id); err != nil {
			t.Fatalf("ValidateStringID(%q) = %v", id, err)
		}
	}
	for _, id := range []string{"", "a%2Fb", "%"} {
		if err := ValidateStringID(id); !errors.Is(err, ErrInvalidStringID) {
			t.Fatalf("ValidateStringID(%q) = %v, want ErrInvalidStringID", id, err)
		}
	}
	if got, want := EscapeID("a/b"), EscapeID("a%2Fb"); got != want {
		t.Fatalf("collision fixture changed: EscapeID(\"a/b\")=%q EscapeID(\"a%%2Fb\")=%q", got, want)
	}
}
