package record

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// This file defines the DALgo path grammar once, next to EscapeID, per
// REQ:path-grammar of the spec/features/schema-subcollections-extensions
// Feature in github.com/dal-go/dalgo:
//
//	path          = [ "/" ] segment *( "/" segment )
//	collection    = name
//	id            = composite-key / placeholder / concrete-id
//	composite-key = "{" <any text containing "="> "}"  ; reserved, not supported yet
//	placeholder   = "{" identifier "}"                 ; schema paths only
//	identifier    = ( ALPHA / "_" ) *( ALPHA / DIGIT / "_" )
//	concrete-id   = <EscapeID output>                  ; never begins with "{"
//
// record.Key.String() already emits this grammar for keys built from
// concrete, scalar ids. github.com/dal-go/dalgo's dbschema package (a
// separate lane, not implemented here) builds dbschema.ParseSchemaPath and
// dbschema.SchemaPath on top of exactly the primitives below, rather than
// with its own splitting or escaping rules.

// ErrInvalidPath indicates that a path string does not follow the path
// grammar's segment structure: an empty path, or an empty segment produced
// by a doubled or trailing "/". SplitPath returns it.
var ErrInvalidPath = errors.New("record: invalid path")

// ErrInvalidCollectionName indicates a collection name violates the path
// grammar's rules for names. ValidateCollectionName returns it.
var ErrInvalidCollectionName = errors.New("record: invalid collection name")

// ErrMalformedIDSegment indicates an id segment is not classifiable as a
// concrete id, a placeholder, or a reserved composite-key segment.
// ClassifyIDSegment returns it.
var ErrMalformedIDSegment = errors.New("record: malformed id segment")

// ErrReservedKeySegment indicates an id segment uses the reserved
// multi-field composite-key grammar "{field=value,...}". Parsing that
// grammar into typed values is deferred to a later Feature; today it is
// rejected. ClassifyIDSegment returns it.
//
// The spec requires the equivalent error at the dbschema layer
// (dbschema.ErrReservedKeySegment, in github.com/dal-go/dalgo) to satisfy
// both errors.Is(err, dbschema.ErrReservedKeySegment) and
// errors.Is(err, dal.ErrNotSupported). record cannot import dalgo — dalgo
// imports record, not the reverse — so it cannot itself wrap dal.ErrNotSupported.
// Instead, a caller that needs that combination (dbschema.ParseSchemaPath)
// detects this sentinel with errors.Is and wraps both of its own sentinels
// around it, for example:
//
//	if errors.Is(err, record.ErrReservedKeySegment) {
//	    return fmt.Errorf("%w: %w", dbschema.ErrReservedKeySegment, dal.ErrNotSupported)
//	}
//
// The resulting error satisfies both target checks through Go's multiple-%w
// wrapping, without dbschema.ErrReservedKeySegment or dal.ErrNotSupported
// ever needing to be defined in, or imported by, this package.
var ErrReservedKeySegment = errors.New("record: multi-field key segments are not supported yet")

// SplitPath splits a DALgo path string into its segments. A leading "/" is
// optional and is stripped; it is never part of a returned segment. An empty
// path, or any empty segment (produced by a doubled or trailing "/"), is
// rejected with ErrInvalidPath.
//
// SplitPath performs no escaping/unescaping and no id classification; pair
// it with PathAddressesCollection for the parity rule, ValidateCollectionName
// for segments at collection positions, and ClassifyIDSegment for segments at
// id positions.
func SplitPath(path string) ([]string, error) {
	p := strings.TrimPrefix(path, "/")
	if p == "" {
		return nil, fmt.Errorf("%w: path is empty", ErrInvalidPath)
	}
	segments := strings.Split(p, "/")
	for _, s := range segments {
		if s == "" {
			return nil, fmt.Errorf("%w: %q has an empty segment", ErrInvalidPath, path)
		}
	}
	return segments, nil
}

// PathAddressesCollection reports whether a path with segmentCount segments
// addresses a collection (an odd count) rather than a record (an even
// count), per the path grammar's parity rule: odd positions (1st, 3rd, …)
// are collections, even positions are ids.
func PathAddressesCollection(segmentCount int) bool {
	return segmentCount%2 == 1
}

// ValidateCollectionName reports whether name is a legal collection name:
// non-empty, not "." or "..", with no leading or trailing whitespace, and
// containing no control character and none of the characters
// / \ { } , = %. Collection names are never escaped.
func ValidateCollectionName(name string) error {
	if name == "" {
		return fmt.Errorf("%w: name is empty", ErrInvalidCollectionName)
	}
	if name == "." || name == ".." {
		return fmt.Errorf("%w: name is %q", ErrInvalidCollectionName, name)
	}
	if strings.TrimSpace(name) != name {
		return fmt.Errorf("%w: %q has leading or trailing whitespace", ErrInvalidCollectionName, name)
	}
	for _, r := range name {
		if unicode.IsControl(r) {
			return fmt.Errorf("%w: %q contains a control character", ErrInvalidCollectionName, name)
		}
		switch r {
		case '/', '\\', '{', '}', ',', '=', '%':
			return fmt.Errorf("%w: %q contains %q", ErrInvalidCollectionName, name, string(r))
		}
	}
	return nil
}

// placeholderNameRe is the identifier rule shared with github.com/dal-go/dalgo's
// access.Capture name (access/resource.go's reCaptureName): the same text
// means the same thing in a schema path and in an access pattern.
var placeholderNameRe = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// ValidPlaceholderName reports whether name is a valid schema-path
// placeholder identifier: (ALPHA / "_") *(ALPHA / DIGIT / "_").
func ValidPlaceholderName(name string) bool {
	return placeholderNameRe.MatchString(name)
}

// IDSegmentKind classifies one id position of a DALgo path, as returned by
// ClassifyIDSegment.
type IDSegmentKind int

const (
	// ConcreteIDSegment is a scalar id that scopes to one parent record.
	ConcreteIDSegment IDSegmentKind = iota
	// PlaceholderIDSegment is a "{name}" id that applies to every parent
	// record. It only ever appears in a schema path, never in a data path.
	PlaceholderIDSegment
)

// ClassifyIDSegment classifies a single, still-escaped id path segment (one
// element returned by SplitPath at an id position), in this order:
//
//  1. begins with "{", ends with "}" and contains "=": a reserved
//     multi-field composite key (e.g. "{country=us,state=ca}"). Returns an
//     error satisfying errors.Is(err, ErrReservedKeySegment).
//  2. "{identifier}" (see ValidPlaceholderName): a placeholder. Returns
//     PlaceholderIDSegment and the identifier.
//  3. any other segment beginning with "{": malformed. Returns an error
//     satisfying errors.Is(err, ErrMalformedIDSegment).
//  4. otherwise: a concrete id, unescaped with UnescapeID. A segment
//     containing a raw '{' '}' ',' '=', or an unrecognised '%' escape, is
//     malformed: the returned error satisfies both
//     errors.Is(err, ErrMalformedIDSegment) and
//     errors.Is(err, ErrInvalidStringID).
func ClassifyIDSegment(segment string) (kind IDSegmentKind, id string, err error) {
	if strings.HasPrefix(segment, "{") {
		if strings.HasSuffix(segment, "}") {
			inner := segment[1 : len(segment)-1]
			if strings.ContainsRune(inner, '=') {
				return 0, "", fmt.Errorf("%w: %q", ErrReservedKeySegment, segment)
			}
			if ValidPlaceholderName(inner) {
				return PlaceholderIDSegment, inner, nil
			}
		}
		return 0, "", fmt.Errorf("%w: %q", ErrMalformedIDSegment, segment)
	}
	unescaped, uerr := UnescapeID(segment)
	if uerr != nil {
		return 0, "", fmt.Errorf("%w: %q: %w", ErrMalformedIDSegment, segment, uerr)
	}
	return ConcreteIDSegment, unescaped, nil
}

// idCharsUnescaper reverses idCharsReplacer (see key.go): each two-hex-digit
// escape code EscapeID can produce, mapped back to the raw byte it replaced.
var idCharsUnescaper = map[string]byte{
	"2E": '.',
	"24": '$',
	"23": '#',
	"5B": '[',
	"5D": ']',
	"2F": '/',
	"7B": '{',
	"7D": '}',
	"2C": ',',
	"3D": '=',
}

// UnescapeID is the exact inverse of EscapeID: for any raw id that passes
// ValidateStringID, UnescapeID(EscapeID(id)) == id, nil. It rejects a
// segment EscapeID could never have produced: a raw, un-escaped '{' '}' ','
// or '=' character (these must always appear escaped in valid EscapeID
// output), or a '%' not immediately followed by one of EscapeID's two-hex
// codes. Both failures return an error satisfying
// errors.Is(err, ErrInvalidStringID), the same sentinel ValidateStringID
// uses for the pre-escape direction of the same invariant.
func UnescapeID(id string) (string, error) {
	if !strings.ContainsAny(id, "%{},=") {
		return id, nil
	}
	var b strings.Builder
	b.Grow(len(id))
	for i := 0; i < len(id); {
		switch id[i] {
		case '{', '}', ',', '=':
			return "", fmt.Errorf("%w: raw %q in escaped id %q", ErrInvalidStringID, string(id[i]), id)
		case '%':
			if i+3 > len(id) {
				return "", fmt.Errorf("%w: truncated escape in %q", ErrInvalidStringID, id)
			}
			code := strings.ToUpper(id[i+1 : i+3])
			ch, ok := idCharsUnescaper[code]
			if !ok {
				return "", fmt.Errorf("%w: unknown escape %%%s in %q", ErrInvalidStringID, code, id)
			}
			b.WriteByte(ch)
			i += 3
			continue
		default:
			b.WriteByte(id[i])
		}
		i++
	}
	return b.String(), nil
}
