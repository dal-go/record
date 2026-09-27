package update

import (
	"testing"
)

func TestByFieldName(t *testing.T) {
	// Simple field
	u1 := ByFieldName("title", "Hello")
	if u1.FieldName() != "title" {
		t.Fatalf("expected title, got %s", u1.FieldName())
	}
	if u1.Value() != "Hello" {
		t.Fatalf("expected Hello, got %v", u1.Value())
	}
	if len(u1.FieldPath()) != 0 {
		t.Fatalf("expected empty fieldPath, got %v", u1.FieldPath())
	}

	// Dot separated field path
	u2 := ByFieldName("user.name", "Bob")
	if u2.FieldName() != "" {
		t.Fatalf("expected empty fieldName, got %s", u2.FieldName())
	}
	if len(u2.FieldPath()) != 2 || u2.FieldPath()[0] != "user" || u2.FieldPath()[1] != "name" {
		t.Fatalf("expected [user name], got %v", u2.FieldPath())
	}
	if u2.Value() != "Bob" {
		t.Fatalf("expected Bob, got %v", u2.Value())
	}

	// Panic on empty fieldName
	assertPanic(t, "empty fieldName", func() {
		ByFieldName("", "val")
	})

	// Panic on invalid dot fieldName (e.g. "user..name")
	assertPanic(t, "invalid dot fieldName", func() {
		ByFieldName("user..name", "val")
	})
}

func TestByFieldPath(t *testing.T) {
	// Valid
	u := ByFieldPath(FieldPath{"settings", "theme"}, "dark")
	if len(u.FieldPath()) != 2 || u.FieldPath()[0] != "settings" || u.FieldPath()[1] != "theme" {
		t.Fatalf("expected [settings theme], got %v", u.FieldPath())
	}
	if u.Value() != "dark" {
		t.Fatalf("expected dark, got %v", u.Value())
	}

	// Panic on empty fieldPath
	assertPanic(t, "empty fieldPath", func() {
		ByFieldPath(FieldPath{}, "val")
	})

	// Panic on invalid fieldPath component (e.g. empty component)
	assertPanic(t, "empty component in fieldPath", func() {
		ByFieldPath(FieldPath{"a", ""}, "val")
	})
}

func TestDeleteByFieldName(t *testing.T) {
	u := DeleteByFieldName("temp")
	if u.FieldName() != "temp" {
		t.Fatalf("expected temp, got %s", u.FieldName())
	}
	if u.Value() != DeleteField {
		t.Fatalf("expected DeleteField, got %v", u.Value())
	}

	// Panic on empty
	assertPanic(t, "empty fieldName in DeleteByFieldName", func() {
		DeleteByFieldName("")
	})
}

func TestDeleteByFieldPath(t *testing.T) {
	u := DeleteByFieldPath("settings", "temp")
	if len(u.FieldPath()) != 2 {
		t.Fatalf("expected 2 components, got %v", u.FieldPath())
	}
	if u.Value() != DeleteField {
		t.Fatalf("expected DeleteField, got %v", u.Value())
	}

	// Panic on empty
	assertPanic(t, "empty path in DeleteByFieldPath", func() {
		DeleteByFieldPath()
	})
}

func TestUpdate_Validate(t *testing.T) {
	// Both empty
	uEmpty := update{}
	if err := uEmpty.Validate(); err == nil {
		t.Fatal("expected error when both empty")
	}

	// fieldName contains dot
	uDot := update{fieldName: "a.b"}
	if err := uDot.Validate(); err == nil {
		t.Fatal("expected error when fieldName contains dot")
	}

	// Both fieldName and fieldPath provided
	uBoth := update{fieldName: "name", fieldPath: FieldPath{"a", "b"}}
	if err := uBoth.Validate(); err == nil {
		t.Fatal("expected error when both fieldName and fieldPath provided")
	}

	// fieldPath with space component
	uSpacePath := update{fieldPath: FieldPath{"a", "   "}}
	if err := uSpacePath.Validate(); err == nil {
		t.Fatal("expected error when fieldPath contains empty/space component")
	}

	// Valid fieldName
	uValidName := update{fieldName: "title"}
	if err := uValidName.Validate(); err != nil {
		t.Fatalf("expected valid, got %v", err)
	}

	// Valid fieldPath
	uValidPath := update{fieldPath: FieldPath{"user", "email"}}
	if err := uValidPath.Validate(); err != nil {
		t.Fatalf("expected valid, got %v", err)
	}
}

func assertPanic(t *testing.T, msg string, f func()) {
	t.Helper()
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic for [%s], but did not panic", msg)
		}
	}()
	f()
}
