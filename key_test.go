package record

import (
	"errors"
	"reflect"
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
