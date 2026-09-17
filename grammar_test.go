package record

import (
	"errors"
	"strings"
	"testing"
)

// --- EscapeID / UnescapeID round trip (AC:id-escaping-extended) ---

func TestEscapeIDEscapesReservedGrammarChars(t *testing.T) {
	k := NewKeyWithID("projects", "{a=b,c}")
	got, want := k.String(), "projects/%7Ba%3Db%2Cc%7D"
	if got != want {
		t.Fatalf("Key.String() = %q, want %q", got, want)
	}
	idSegment := strings.TrimPrefix(got, "projects/")
	unescaped, err := UnescapeID(idSegment)
	if err != nil {
		t.Fatalf("UnescapeID(%q) error = %v", idSegment, err)
	}
	if unescaped != "{a=b,c}" {
		t.Fatalf("UnescapeID(%q) = %q, want %q", idSegment, unescaped, "{a=b,c}")
	}
	if got, want := NewKeyWithID("projects", "a.b").String(), "projects/a%2Eb"; got != want {
		t.Fatalf("Key.String() = %q, want %q (existing escapes must still work)", got, want)
	}
}

func TestEscapeIDUnescapeIDRoundTrip(t *testing.T) {
	ids := []string{
		"plain",
		"a.b",
		"a$b",
		"a#b",
		"a[b]",
		"a/b",
		"a{b}",
		"a,b",
		"a=b",
		"{a=b,c}",
		"base64==",
		".$#[]/{},=",
		"",
	}
	for _, id := range ids {
		escaped := EscapeID(id)
		got, err := UnescapeID(escaped)
		if err != nil {
			t.Fatalf("UnescapeID(EscapeID(%q)=%q) error = %v", id, escaped, err)
		}
		if got != id {
			t.Fatalf("UnescapeID(EscapeID(%q)) = %q, want %q", id, got, id)
		}
	}
}

func TestUnescapeIDRejectsMalformedInput(t *testing.T) {
	for _, id := range []string{
		"a{b",
		"a}b",
		"a,b",
		"a=b",
		"a%",
		"a%2",
		"a%tug",
		"a%ZZ",
	} {
		if _, err := UnescapeID(id); !errors.Is(err, ErrInvalidStringID) {
			t.Fatalf("UnescapeID(%q) error = %v, want ErrInvalidStringID", id, err)
		}
	}
}

// --- ValidPlaceholderName ---

func TestValidPlaceholderName(t *testing.T) {
	for _, name := range []string{"projectID", "_x1", "a", "A_1", "_"} {
		if !ValidPlaceholderName(name) {
			t.Fatalf("ValidPlaceholderName(%q) = false, want true", name)
		}
	}
	for _, name := range []string{"", "1x", "ext-id", "ext.id", "ext id", "ext/id", "ext=id"} {
		if ValidPlaceholderName(name) {
			t.Fatalf("ValidPlaceholderName(%q) = true, want false", name)
		}
	}
}

// --- ValidateCollectionName ---

func TestValidateCollectionName(t *testing.T) {
	for _, name := range []string{"users", "ext", "queries2", "a_b", "a-b"} {
		if err := ValidateCollectionName(name); err != nil {
			t.Fatalf("ValidateCollectionName(%q) = %v, want nil", name, err)
		}
	}
	cases := []string{
		"",
		".",
		"..",
		" ext",
		"ext ",
		"ext/id",
		"ext\\id",
		"ext{id",
		"ext}id",
		"ext,id",
		"ext=id",
		"ext%id",
		"ext\tid",
	}
	for _, name := range cases {
		if err := ValidateCollectionName(name); !errors.Is(err, ErrInvalidCollectionName) {
			t.Fatalf("ValidateCollectionName(%q) = %v, want ErrInvalidCollectionName", name, err)
		}
	}
}

// --- SplitPath ---

func TestSplitPathAcceptsOptionalLeadingSlash(t *testing.T) {
	cases := []struct {
		path string
		want []string
	}{
		{"users", []string{"users"}},
		{"/users", []string{"users"}},
		{"ext/datatug/projects", []string{"ext", "datatug", "projects"}},
		{"/ext/datatug/projects/{projectID}/queries", []string{"ext", "datatug", "projects", "{projectID}", "queries"}},
	}
	for _, c := range cases {
		got, err := SplitPath(c.path)
		if err != nil {
			t.Fatalf("SplitPath(%q) error = %v", c.path, err)
		}
		if len(got) != len(c.want) {
			t.Fatalf("SplitPath(%q) = %v, want %v", c.path, got, c.want)
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Fatalf("SplitPath(%q) = %v, want %v", c.path, got, c.want)
			}
		}
	}
}

func TestSplitPathRejectsMalformedPaths(t *testing.T) {
	for _, path := range []string{"", "/", "ext/datatug/", "ext//projects", "//"} {
		if _, err := SplitPath(path); !errors.Is(err, ErrInvalidPath) {
			t.Fatalf("SplitPath(%q) error = %v, want ErrInvalidPath", path, err)
		}
	}
}

// --- PathAddressesCollection ---

func TestPathAddressesCollection(t *testing.T) {
	for count, want := range map[int]bool{1: true, 3: true, 5: true, 2: false, 4: false, 0: false} {
		if got := PathAddressesCollection(count); got != want {
			t.Fatalf("PathAddressesCollection(%d) = %v, want %v", count, got, want)
		}
	}
}

// --- ClassifyIDSegment ---

func TestClassifyIDSegmentConcreteID(t *testing.T) {
	kind, id, err := ClassifyIDSegment("a%2Fb")
	if err != nil {
		t.Fatalf("ClassifyIDSegment(%q) error = %v", "a%2Fb", err)
	}
	if kind != ConcreteIDSegment {
		t.Fatalf("ClassifyIDSegment(%q) kind = %v, want ConcreteIDSegment", "a%2Fb", kind)
	}
	if id != "a/b" {
		t.Fatalf("ClassifyIDSegment(%q) id = %q, want %q", "a%2Fb", id, "a/b")
	}
}

func TestClassifyIDSegmentPlaceholder(t *testing.T) {
	kind, id, err := ClassifyIDSegment("{projectID}")
	if err != nil {
		t.Fatalf("ClassifyIDSegment(%q) error = %v", "{projectID}", err)
	}
	if kind != PlaceholderIDSegment {
		t.Fatalf("ClassifyIDSegment(%q) kind = %v, want PlaceholderIDSegment", "{projectID}", kind)
	}
	if id != "projectID" {
		t.Fatalf("ClassifyIDSegment(%q) id = %q, want %q", "{projectID}", id, "projectID")
	}
}

func TestClassifyIDSegmentReservedCompositeKey(t *testing.T) {
	for _, segment := range []string{"{country=us,state=ca}", "{country=us}", "{a=b,c}"} {
		_, _, err := ClassifyIDSegment(segment)
		if !errors.Is(err, ErrReservedKeySegment) {
			t.Fatalf("ClassifyIDSegment(%q) error = %v, want ErrReservedKeySegment", segment, err)
		}
	}
}

func TestClassifyIDSegmentMalformed(t *testing.T) {
	for _, segment := range []string{"{}", "{ext-id}", "{extID", "data,tug", "data%tug", "data%2"} {
		_, _, err := ClassifyIDSegment(segment)
		if !errors.Is(err, ErrMalformedIDSegment) {
			t.Fatalf("ClassifyIDSegment(%q) error = %v, want ErrMalformedIDSegment", segment, err)
		}
	}
}

// --- Composed coverage of AC:malformed-and-record-paths-rejected and
// AC:multi-field-key-segment-reserved, exercised entirely through the
// primitives this package exports. dbschema.ParseSchemaPath (a separate
// lane, out of scope here) builds its own exported parser on top of exactly
// these primitives, per REQ:path-grammar.

// classifyPathForTest is a minimal, unexported stand-in for the future
// dbschema.ParseSchemaPath, composed only from this package's exported
// grammar primitives. It exists to prove those primitives are sufficient to
// implement every example in the spec's ACs; it is not part of the public
// API.
func classifyPathForTest(path string) error {
	segments, err := SplitPath(path)
	if err != nil {
		return err
	}
	if !PathAddressesCollection(len(segments)) {
		return errRecordPathNotCollectionForTest
	}
	seenPlaceholders := map[string]bool{}
	for i, segment := range segments {
		if i%2 == 0 {
			if err := ValidateCollectionName(segment); err != nil {
				return err
			}
			continue
		}
		kind, id, err := ClassifyIDSegment(segment)
		if err != nil {
			return err
		}
		if kind == PlaceholderIDSegment {
			if seenPlaceholders[id] {
				return errDuplicatePlaceholderForTest
			}
			seenPlaceholders[id] = true
		}
	}
	return nil
}

var (
	errRecordPathNotCollectionForTest = errors.New("test: path addresses a record, not a collection")
	errDuplicatePlaceholderForTest    = errors.New("test: duplicate placeholder name")
)

func TestMalformedAndRecordPathsRejected(t *testing.T) {
	evenCount := []string{"ext/datatug", "projects/{projectID}"}
	for _, path := range evenCount {
		err := classifyPathForTest(path)
		if !errors.Is(err, errRecordPathNotCollectionForTest) {
			t.Fatalf("classifyPathForTest(%q) error = %v, want record-path-not-collection", path, err)
		}
	}

	malformed := []string{
		"",
		"/",
		"ext/datatug/",
		"ext//projects",
		"ext/{}/projects",
		"ext/{ext-id}/projects",
		"ext/{extID/projects",
		"ext/{id}/projects/{id}/queries",
		"ext/datatug/proj{ects",
		"ext/data,tug/projects",
		"ext/data%tug/projects",
	}
	for _, path := range malformed {
		err := classifyPathForTest(path)
		if err == nil {
			t.Fatalf("classifyPathForTest(%q) = nil, want an error", path)
		}
		if errors.Is(err, errRecordPathNotCollectionForTest) {
			t.Fatalf("classifyPathForTest(%q) = record-path-not-collection, want a malformed-path error", path)
		}
	}
}

func TestMultiFieldKeySegmentReserved(t *testing.T) {
	for _, path := range []string{"countries/{country=us,state=ca}/cities", "/countries/{country=us}/cities"} {
		err := classifyPathForTest(path)
		if !errors.Is(err, ErrReservedKeySegment) {
			t.Fatalf("classifyPathForTest(%q) error = %v, want ErrReservedKeySegment", path, err)
		}
	}
}
