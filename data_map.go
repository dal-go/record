package record

import (
	"encoding/json"
	"fmt"
	"maps"
	"reflect"
	"strings"
)

// DataToMap converts record data into a map keyed by field name. A db tag takes
// precedence over a json tag at the top level.
func DataToMap(data any) (map[string]any, error) {
	if data == nil {
		return nil, nil
	}
	if m, ok := data.(map[string]any); ok {
		return m, nil
	}
	b, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("record: marshal data of type %T: %w", data, err)
	}
	m := map[string]any{}
	if err = json.Unmarshal(b, &m); err != nil {
		return nil, fmt.Errorf("record: convert data of type %T to map: %w", data, err)
	}
	for _, r := range fieldTagRenames(reflect.TypeOf(data)) {
		if v, ok := m[r.from]; ok {
			m[r.to] = v
			delete(m, r.from)
		}
	}
	return m, nil
}

// MapToData populates target from src. A db tag takes precedence over a json
// tag at the top level.
func MapToData(target any, src map[string]any) error {
	if m, ok := target.(map[string]any); ok {
		maps.Copy(m, src)
		return nil
	}
	remapped := maps.Clone(src)
	if remapped == nil {
		remapped = map[string]any{}
	}
	for _, r := range fieldTagRenames(reflect.TypeOf(target)) {
		if v, ok := remapped[r.to]; ok {
			remapped[r.from] = v
			delete(remapped, r.to)
		}
	}
	b, err := json.Marshal(remapped)
	if err != nil {
		return fmt.Errorf("record: marshal data map: %w", err)
	}
	if err = json.Unmarshal(b, target); err != nil {
		return fmt.Errorf("record: unmarshal data into %T: %w", target, err)
	}
	return nil
}

type tagRename struct{ from, to string }

func fieldTagRenames(t reflect.Type) []tagRename {
	for t != nil && (t.Kind() == reflect.Pointer || t.Kind() == reflect.Interface) {
		t = t.Elem()
	}
	if t == nil || t.Kind() != reflect.Struct {
		return nil
	}
	var renames []tagRename
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.Anonymous {
			renames = append(renames, fieldTagRenames(f.Type)...)
			continue
		}
		if f.PkgPath != "" {
			continue
		}
		jsonKey := tagName(f.Tag.Get("json"), f.Name)
		if jsonKey == "" {
			continue
		}
		dbKey := tagName(f.Tag.Get("db"), "")
		if dbKey == "" || dbKey == jsonKey {
			continue
		}
		renames = append(renames, tagRename{from: jsonKey, to: dbKey})
	}
	return renames
}

func tagName(tag, fallback string) string {
	if tag == "" {
		return fallback
	}
	name, _, _ := strings.Cut(tag, ",")
	if name == "-" {
		return ""
	}
	if name == "" {
		return fallback
	}
	return name
}
