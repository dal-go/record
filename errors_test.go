package record

import (
	"errors"
	"fmt"
	"testing"
)

func TestIsNotFound(t *testing.T) {
	if IsNotFound(nil) {
		t.Fatal("IsNotFound(nil) = true, want false")
	}
	if !IsNotFound(ErrRecordNotFound) {
		t.Fatal("IsNotFound(ErrRecordNotFound) = false, want true")
	}
	wrapped := fmt.Errorf("get %s: %w", "users/1", ErrRecordNotFound)
	if !IsNotFound(wrapped) {
		t.Fatalf("IsNotFound(%v) = false, want true", wrapped)
	}
	if IsNotFound(ErrRecordExists) {
		t.Fatal("IsNotFound(ErrRecordExists) = true, want false")
	}
	if IsNotFound(errors.New("some other failure")) {
		t.Fatal("IsNotFound(unrelated error) = true, want false")
	}
}

func TestIsAlreadyExists(t *testing.T) {
	if IsAlreadyExists(nil) {
		t.Fatal("IsAlreadyExists(nil) = true, want false")
	}
	if !IsAlreadyExists(ErrRecordExists) {
		t.Fatal("IsAlreadyExists(ErrRecordExists) = false, want true")
	}
	wrapped := fmt.Errorf("insert %s: %w", "users/1", ErrRecordExists)
	if !IsAlreadyExists(wrapped) {
		t.Fatalf("IsAlreadyExists(%v) = false, want true", wrapped)
	}
	if IsAlreadyExists(ErrRecordNotFound) {
		t.Fatal("IsAlreadyExists(ErrRecordNotFound) = true, want false")
	}
	if IsAlreadyExists(errors.New("some other failure")) {
		t.Fatal("IsAlreadyExists(unrelated error) = true, want false")
	}
}
