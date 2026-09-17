package record

import (
	"errors"
	"fmt"
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
// AC:key-constructors-keep-incomplete-keys.
func TestNewKeyWithIDPanicsOnInvalidStringID(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("NewKeyWithID did not panic on a string id containing '%'")
		}
		msg := fmt.Sprint(r)
		if !strings.Contains(msg, ErrInvalidStringID.Error()) {
			t.Fatalf("panic message = %q, want it to name %v", msg, ErrInvalidStringID)
		}
	}()
	NewKeyWithID("users", "a%2Fb")
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
