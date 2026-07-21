package record

import (
	"testing"

	"github.com/dal-go/record/update"
)

func TestDataWithIDSharesEnvelopeData(t *testing.T) {
	type user struct{ Name string }
	data := &user{Name: "Ada"}
	v := NewDataWithID("u1", NewKeyWithID("users", "u1"), data)
	v.Record.SetError(nil)
	if v.Record.Data() != data || v.Data != data {
		t.Fatal("typed and untyped data must reference the same value")
	}
}

func TestChangesQueuesCommandsAndResets(t *testing.T) {
	key := NewKeyWithID("users", "u1")
	rec := NewRecordWithData(key, &struct{}{}).SetError(nil)
	changes := &Changes{}
	changes.QueueForInsert(rec)
	changes.RecordsToUpdate = []*Updates{{Record: rec, Updates: []update.Update{update.ByFieldName("name", "Ada")}}}
	changes.RecordsToDelete = []*Key{key}

	if len(changes.RecordsToInsert()) != 1 {
		t.Fatal("expected queued insert")
	}
	changes.Reset()
	if len(changes.RecordsToInsert()) != 0 || len(changes.RecordsToUpdate) != 0 || len(changes.RecordsToDelete) != 0 {
		t.Fatal("Reset must clear every command queue")
	}
}

func TestDataMapRoundTripHonorsDBTag(t *testing.T) {
	type user struct {
		Name string `json:"name" db:"display_name"`
	}
	m, err := DataToMap(&user{Name: "Ada"})
	if err != nil {
		t.Fatal(err)
	}
	if got := m["display_name"]; got != "Ada" {
		t.Fatalf("DataToMap() = %#v", m)
	}

	var decoded user
	if err := MapToData(&decoded, m); err != nil {
		t.Fatal(err)
	}
	if decoded.Name != "Ada" {
		t.Fatalf("MapToData() = %#v", decoded)
	}
}
