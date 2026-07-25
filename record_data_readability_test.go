package record

import (
	"reflect"
	"testing"
)

// The invariant these tests pin:
//
//	a record whose data the CALLER SUPPLIED is readable immediately, because
//	there is no retrieval to wait for;
//	a record that is an EMPTY ENVELOPE awaiting a load is not readable until a
//	loader has said how the load went.
//
// Data()'s panic is deliberate — it stops data being read before anyone has
// checked the error. That guard is only correct if constructors classify their
// records correctly, and two of the four did not: NewRecordWithData and
// NewRecordWithoutKey left the error unset even though the caller had handed
// over the data, so reading it panicked. NewRecordWithIncompleteKey had it
// right all along, which is what showed the other two were wrong.
//
// This mattered beyond ergonomics: dalgo's own dal.BeforeSave validation hook
// reads Data() on a record about to be inserted, so it could not work at all
// until this was fixed — see github.com/dal-go/dalgo.
//
// Note these tests assert BEHAVIOUR (does Data() panic?) rather than the
// internal err field. Error() deliberately maps the ErrNoError sentinel to nil,
// so it reports nil for BOTH a supplied record and an unloaded one and cannot
// tell them apart. Data() is the only observable difference.

func mustNotPanic(t *testing.T, name string, f func() any) any {
	t.Helper()
	var got any
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("%s: Data() panicked on a record whose data the caller supplied: %v", name, r)
			}
		}()
		got = f()
	}()
	return got
}

func TestSuppliedDataIsReadableImmediately(t *testing.T) {
	type payload struct{ Name string }
	key := NewKeyWithID("things", "thing1")

	t.Run("NewRecordWithData", func(t *testing.T) {
		data := &payload{Name: "supplied"}
		r := NewRecordWithData(key, data)

		if err := r.Error(); err != nil {
			t.Fatalf("Error() = %v, want nil", err)
		}
		got := mustNotPanic(t, "NewRecordWithData", r.Data)
		if got != any(data) {
			t.Errorf("Data() = %#v, want the supplied value back", got)
		}
	})

	t.Run("NewRecordWithoutKey", func(t *testing.T) {
		data := &payload{Name: "keyless"}
		r := NewRecordWithoutKey(data)

		if err := r.Error(); err != nil {
			t.Fatalf("Error() = %v, want nil", err)
		}
		if got := mustNotPanic(t, "NewRecordWithoutKey", r.Data); got != any(data) {
			t.Errorf("Data() = %#v, want the supplied value back", got)
		}
	})

	t.Run("NewRecordWithIncompleteKey was already correct", func(t *testing.T) {
		data := &payload{Name: "incomplete"}
		r := NewRecordWithIncompleteKey("things", reflect.String, data)

		if err := r.Error(); err != nil {
			t.Fatalf("Error() = %v, want nil", err)
		}
		if got := mustNotPanic(t, "NewRecordWithIncompleteKey", r.Data); got != any(data) {
			t.Errorf("Data() = %#v, want the supplied value back", got)
		}
	})
}

// TestEmptyEnvelopeStillGuardsItsData is the other half of the invariant, and
// the reason this fix is narrow rather than "stop panicking". NewRecord builds
// an envelope for a load that has not happened, so reading it must still panic
// — otherwise the guard would be gone and callers could read stale or absent
// data without checking the error.
func TestEmptyEnvelopeStillGuardsItsData(t *testing.T) {
	r := NewRecord(NewKeyWithID("things", "thing1"))

	if r.Error() != nil {
		t.Fatalf("Error() = %v, want nil — nothing has loaded this record yet", r.Error())
	}

	defer func() {
		if recover() == nil {
			t.Fatal("Data() did not panic on a record awaiting a load — the guard has been lost")
		}
	}()
	_ = r.Data()
}
