package record

import (
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
