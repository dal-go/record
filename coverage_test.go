package record

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

type mockRecord struct {
	key     *Key
	data    any
	err     error
	changed bool
}

func (m *mockRecord) Key() *Key         { return m.key }
func (m *mockRecord) Error() error      { return m.err }
func (m *mockRecord) Exists() bool      { return m.err == nil }
func (m *mockRecord) SetError(err error) Record {
	m.err = err
	return m
}
func (m *mockRecord) Data() any         { return m.data }
func (m *mockRecord) HasChanged() bool  { return m.changed }
func (m *mockRecord) MarkAsChanged()    { m.changed = true }

func TestChanges_Coverage(t *testing.T) {
	c := &Changes{}

	// Panic on nil record
	assertPanic(t, "nil record", func() {
		c.QueueForInsert(nil)
	})

	// Panic on record with nil key
	assertPanic(t, "nil record key", func() {
		c.QueueForInsert(&mockRecord{})
	})

	// Panic on record with nil data
	assertPanic(t, "nil record data", func() {
		c.QueueForInsert(&mockRecord{key: NewKeyWithID("col", "id1")})
	})

	// Queue valid record
	k1 := NewKeyWithID("col", "id1")
	r1 := NewRecordWithData(k1, &struct{}{})
	c.QueueForInsert(r1)

	// Panic on already queued record
	assertPanic(t, "already queued", func() {
		c.QueueForInsert(r1)
	})

	// RecordsToInsert
	recs := c.RecordsToInsert()
	if len(recs) != 1 || !recs[0].Key().Equal(k1) {
		t.Fatalf("unexpected RecordsToInsert: %v", recs)
	}

	// Reset
	c.Reset()
	if len(c.RecordsToInsert()) != 0 {
		t.Fatalf("expected 0 after Reset, got %d", len(c.RecordsToInsert()))
	}
}

type embeddedStruct struct {
	Inner string `json:"inner" db:"inner_db"`
}

type sampleStruct struct {
	embeddedStruct
	unexported int
	Skipped    string `json:"-"`
	Title      string `json:"title" db:"db_title"`
	SameTag    string `json:"same" db:"same"`
	NoDbTag    string `json:"no_db"`
	EmptyDbTag string `json:"empty_db" db:",omitempty"`
}

func TestDataToMap_And_MapToData_Coverage(t *testing.T) {
	// DataToMap nil
	m1, err := DataToMap(nil)
	if err != nil || m1 != nil {
		t.Fatalf("expected nil, nil, got %v, %v", m1, err)
	}

	// DataToMap already map[string]any
	rawMap := map[string]any{"k": "v"}
	m2, err := DataToMap(rawMap)
	if err != nil || m2["k"] != "v" {
		t.Fatalf("expected rawMap, got %v, %v", m2, err)
	}

	// DataToMap marshal error
	if _, err := DataToMap(make(chan int)); err == nil {
		t.Fatal("expected error on unmarshalable type")
	}

	// DataToMap unmarshal to map error (e.g. integer)
	if _, err := DataToMap(12345); err == nil {
		t.Fatal("expected error when converting int to map")
	}

	// DataToMap with tag renames
	s := sampleStruct{
		embeddedStruct: embeddedStruct{Inner: "in"},
		unexported:     10,
		Skipped:        "skip",
		Title:          "hello",
		SameTag:        "same_val",
		NoDbTag:        "no_db_val",
		EmptyDbTag:     "empty_db_val",
	}
	m3, err := DataToMap(s)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m3["db_title"] != "hello" || m3["inner_db"] != "in" {
		t.Fatalf("expected db tags in map, got: %v", m3)
	}
	if _, ok := m3["title"]; ok {
		t.Fatal("expected json title to be deleted from map")
	}

	// MapToData with map target
	targetMap := map[string]any{}
	if err := MapToData(targetMap, map[string]any{"x": 1}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if targetMap["x"] != 1 {
		t.Fatalf("expected x=1 in targetMap")
	}

	// MapToData with nil src
	targetStruct := sampleStruct{}
	if err := MapToData(&targetStruct, nil); err != nil {
		t.Fatalf("unexpected error on nil src: %v", err)
	}

	// MapToData with db tags remapped back to json tags
	srcMap := map[string]any{
		"db_title": "remap_hello",
		"inner_db": "remap_inner",
	}
	if err := MapToData(&targetStruct, srcMap); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if targetStruct.Title != "remap_hello" || targetStruct.Inner != "remap_inner" {
		t.Fatalf("unexpected remapped struct: %+v", targetStruct)
	}

	// MapToData marshal error
	if err := MapToData(&sampleStruct{}, map[string]any{"bad": make(chan int)}); err == nil {
		t.Fatal("expected marshal error in MapToData")
	}

	// MapToData unmarshal error
	type numTarget struct {
		Val int `json:"val"`
	}
	nt := numTarget{}
	if err := MapToData(&nt, map[string]any{"val": "not_an_int"}); err == nil {
		t.Fatal("expected unmarshal error in MapToData")
	}

	// fieldTagRenames on non-struct or nil
	if len(fieldTagRenames(nil)) != 0 {
		t.Fatal("expected empty renames for nil type")
	}
	if len(fieldTagRenames(reflect.TypeOf(123))) != 0 {
		t.Fatal("expected empty renames for int type")
	}

	// tagName
	if tagName("", "default") != "default" {
		t.Fatalf("expected default, got %s", tagName("", "default"))
	}
	if tagName("-", "default") != "" {
		t.Fatalf("expected empty for -, got %s", tagName("-", "default"))
	}
	if tagName(",omitempty", "default") != "default" {
		t.Fatalf("expected default for empty leading part, got %s", tagName(",omitempty", "default"))
	}
	if tagName("custom,omitempty", "default") != "custom" {
		t.Fatalf("expected custom, got %s", tagName("custom,omitempty", "default"))
	}
}

func TestDataWithID_Panics(t *testing.T) {
	k := NewKeyWithID("col", "id1")

	// Nil key
	assertPanic(t, "nil key", func() {
		NewDataWithID(1, nil, &struct{}{})
	})

	// Nil pointer data
	assertPanic(t, "nil pointer data", func() {
		var nilPtr *sampleStruct
		NewDataWithID(1, k, nilPtr)
	})

	// Pointer to non-struct/non-map
	assertPanic(t, "pointer to int", func() {
		val := 10
		NewDataWithID(1, k, &val)
	})

	// Nil interface
	assertPanic(t, "nil interface", func() {
		NewDataWithID[int, any](1, k, nil)
	})

	// Non-pointer/non-interface
	assertPanic(t, "int value", func() {
		NewDataWithID(1, k, 123)
	})

	// Valid struct pointer
	d1 := NewDataWithID(1, k, &sampleStruct{Title: "ok"})
	if d1.Data.Title != "ok" || d1.ID != 1 {
		t.Fatalf("unexpected DataWithID: %+v", d1)
	}

	// Valid map pointer
	m := map[string]string{"a": "b"}
	d2 := NewDataWithID("m1", k, &m)
	if (*d2.Data)["a"] != "b" {
		t.Fatalf("unexpected DataWithID map: %+v", d2)
	}
}

func TestFieldVal_Coverage(t *testing.T) {
	// Empty name
	fv1 := FieldVal{Name: ""}
	if err := fv1.Validate(); err == nil || !strings.Contains(err.Error(), "missing field name") {
		t.Fatalf("expected missing field name error, got %v", err)
	}
	// Whitespace name
	fv2 := FieldVal{Name: "   "}
	if err := fv2.Validate(); err == nil {
		t.Fatal("expected error for whitespace field name")
	}
	// Valid
	fv3 := FieldVal{Name: "id", Value: 123}
	if err := fv3.Validate(); err != nil {
		t.Fatalf("expected valid, got %v", err)
	}
}

type validatingID struct {
	val string
}

func (v validatingID) Validate() error {
	if v.val == "invalid" {
		return errors.New("invalid custom ID")
	}
	return nil
}

func TestKey_Coverage(t *testing.T) {
	// String() on invalid key panics
	kInvalid := &Key{collection: ""}
	assertPanic(t, "String() on invalid key", func() {
		_ = kInvalid.String()
	})

	// CollectionPath() panics on empty collection
	assertPanic(t, "CollectionPath() on empty collection", func() {
		_ = kInvalid.CollectionPath()
	})

	// CollectionPath() valid hierarchical key
	kP := NewKeyWithID("parents", "p1")
	kC := NewKeyWithParentAndID(kP, "children", "c1")
	if cp := kC.CollectionPath(); cp != "parents/children" {
		t.Fatalf("expected parents/children, got %s", cp)
	}

	// reverseStringsJoin empty
	if s := reverseStringsJoin(nil, "/"); s != "" {
		t.Fatalf("expected empty string, got %s", s)
	}

	// Validate() with empty collection
	if err := kInvalid.Validate(); err == nil {
		t.Fatal("expected error on empty collection")
	}

	// Validate() with invalid FieldVal slice
	kFieldsBad := NewKeyWithFields("col", FieldVal{Name: ""})
	if err := kFieldsBad.Validate(); err == nil {
		t.Fatal("expected error on invalid FieldVal")
	}

	// Validate() with validating ID
	kValidatingBad := NewKeyWithID("col", validatingID{val: "invalid"})
	if err := kValidatingBad.Validate(); err == nil {
		t.Fatal("expected error on validatingID")
	}
	kValidatingGood := NewKeyWithID("col", validatingID{val: "good"})
	if err := kValidatingGood.Validate(); err != nil {
		t.Fatalf("expected valid, got %v", err)
	}

	// NewKeyWithID with empty collection panics
	assertPanic(t, "NewKeyWithID empty collection", func() {
		NewKeyWithID("", "id1")
	})

	// NewIncompleteKey with Invalid kind panics
	assertPanic(t, "NewIncompleteKey Invalid kind", func() {
		NewIncompleteKey("col", reflect.Invalid, nil)
	})

	// WithFields
	kWithFields, err := NewKeyWithOptions("col", WithFields([]FieldVal{{Name: "f1", Value: "v1"}}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(kWithFields.ID.([]FieldVal)) != 1 {
		t.Fatalf("expected 1 field")
	}

	// WithParentKey nil panics
	assertPanic(t, "WithParentKey nil", func() {
		WithParentKey(nil)
	})

	// WithStringID & WithIntID
	kStr, err := NewKeyWithOptions("col", WithStringID("sid"))
	if err != nil || kStr.ID != "sid" {
		t.Fatalf("unexpected WithStringID: %v, %v", kStr, err)
	}
	kInt, err := NewKeyWithOptions("col", WithIntID(99))
	if err != nil || kInt.ID != 99 {
		t.Fatalf("unexpected WithIntID: %v, %v", kInt, err)
	}

	// NewKeyWithOptions with empty collection returns error
	if _, err := NewKeyWithOptions(""); err == nil {
		t.Fatal("expected error for empty collection in NewKeyWithOptions")
	}

	// NewKeyWithOptions error option
	optErr := func(k *Key) error { return errors.New("option failure") }
	if _, err := NewKeyWithOptions("col", optErr); err == nil {
		t.Fatal("expected error from failed option")
	}

	// EqualKeys coverage
	if !EqualKeys(nil, nil) {
		t.Fatal("expected nil == nil")
	}
	kA := NewKeyWithID("col", "id1")
	if EqualKeys(nil, kA) || EqualKeys(kA, nil) {
		t.Fatal("expected nil != non-nil")
	}

	// EqualKeys with circular keys panics
	assertPanic(t, "circular key", func() {
		kParent := NewKeyWithID("col", "c1")
		kChild := NewKeyWithParentAndID(kParent, "col", "c1") // same collection & ID as parent!
		EqualKeys(kChild, kChild)
	})

	// EqualKeys with differing parent collections or IDs
	kP1 := NewKeyWithID("parentCol1", "p1")
	kP2 := NewKeyWithID("parentCol2", "p1")
	kC1 := NewKeyWithParentAndID(kP1, "sub", "1")
	kC2 := NewKeyWithParentAndID(kP2, "sub", "1")
	if EqualKeys(kC1, kC2) {
		t.Fatal("expected false for different parent collections")
	}

	kP3 := NewKeyWithID("parentCol1", "diffID")
	kC3 := NewKeyWithParentAndID(kP3, "sub", "1")
	if EqualKeys(kC1, kC3) {
		t.Fatal("expected false for different parent IDs")
	}
}

func TestRecord_Coverage(t *testing.T) {
	k := NewKeyWithID("users", "u1")

	// Exists panics before SetError
	r := NewRecord(k)
	assertPanic(t, "Exists() before SetError", func() {
		_ = r.Exists()
	})

	// Exists with generic error panics
	rErr := NewRecord(k).SetError(errors.New("db crash"))
	assertPanic(t, "Exists() with error", func() {
		_ = rErr.Exists()
	})

	// Exists with ErrNoError returns true
	rGood := NewRecord(k).SetError(nil)
	if !rGood.Exists() {
		t.Fatal("expected Exists() == true")
	}

	// Exists with ErrRecordNotFound returns false
	rNotFound := NewRecord(k).SetError(ErrRecordNotFound)
	if rNotFound.Exists() {
		t.Fatal("expected Exists() == false")
	}

	// HasChanged & MarkAsChanged
	if r.HasChanged() {
		t.Fatal("expected HasChanged == false initially")
	}
	r.MarkAsChanged()
	if !r.HasChanged() {
		t.Fatal("expected HasChanged == true after MarkAsChanged")
	}

	// Data() panics before SetError
	rNoDataLoad := NewRecord(k)
	assertPanic(t, "Data() before SetError", func() {
		_ = rNoDataLoad.Data()
	})

	// Data() panics on general error
	assertPanic(t, "Data() with general error", func() {
		_ = rErr.Data()
	})

	// Data() returns data on ErrRecordNotFound
	rNotFoundWithData := NewRecordWithData(k, "placeholder").SetError(ErrRecordNotFound)
	if rNotFoundWithData.Data() != "placeholder" {
		t.Fatalf("expected placeholder, got %v", rNotFoundWithData.Data())
	}

	// Error() returns nil on ErrNoRecord or nil
	if rGood.Error() != nil {
		t.Fatalf("expected nil error, got %v", rGood.Error())
	}
	if rNotFound.Error() != nil {
		t.Fatalf("expected nil error on not found, got %v", rNotFound.Error())
	}
	if rErr.Error() == nil || rErr.Error().Error() != "db crash" {
		t.Fatalf("expected db crash error, got %v", rErr.Error())
	}

	// newRecordWithOnlyKey panics on nil key or invalid key
	assertPanic(t, "nil key in newRecordWithOnlyKey", func() {
		NewRecord(nil)
	})
	assertPanic(t, "invalid key in newRecordWithOnlyKey", func() {
		NewRecord(&Key{})
	})

	// NewRecordWithoutKey panics:
	// nil data
	assertPanic(t, "nil data in NewRecordWithoutKey", func() {
		NewRecordWithoutKey(nil)
	})
	// map non-string key
	assertPanic(t, "map non-string key", func() {
		NewRecordWithoutKey(map[int]string{1: "a"})
	})
	// pointer to map
	assertPanic(t, "pointer to map", func() {
		m := map[string]string{}
		NewRecordWithoutKey(&m)
	})
	// pointer to slice
	assertPanic(t, "pointer to slice", func() {
		sl := []string{}
		NewRecordWithoutKey(&sl)
	})
	// primitive int
	assertPanic(t, "primitive int", func() {
		NewRecordWithoutKey(42)
	})
	// valid map
	rMap := NewRecordWithoutKey(map[string]string{"key": "val"})
	if rMap.Data() == nil {
		t.Fatal("expected non-nil data")
	}
	// valid slice
	rSlice := NewRecordWithoutKey([]string{"a", "b"})
	if rSlice.Data() == nil {
		t.Fatal("expected non-nil data")
	}

	// AnyRecordWithError
	if err := AnyRecordWithError(rGood, rNotFound); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if err := AnyRecordWithError(rGood, rErr, rNotFound); err == nil || err.Error() != "db crash" {
		t.Fatalf("expected db crash error, got %v", err)
	}
}

func TestWithID_Coverage(t *testing.T) {
	k := NewKeyWithID("col", "id1")

	// FullID is empty
	w1 := WithID[int]{ID: 10, FullID: "", Key: k}
	s1 := w1.String()
	if !strings.Contains(s1, "FullID=nil") {
		t.Fatalf("expected FullID=nil, got %s", s1)
	}

	// String ID with FullID
	w2 := WithID[string]{ID: "id1", FullID: "col/id1", Key: k}
	s2 := w2.String()
	if !strings.Contains(s2, `ID="id1"`) || !strings.Contains(s2, `FullID="col/id1"`) {
		t.Fatalf("unexpected string: %s", s2)
	}

	// Int ID with FullID
	w3 := WithID[int]{ID: 99, FullID: "col/99", Key: k}
	s3 := w3.String()
	if !strings.Contains(s3, `ID=99`) || !strings.Contains(s3, `FullID="col/99"`) {
		t.Fatalf("unexpected string: %s", s3)
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
