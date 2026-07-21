// Package update defines persistence-neutral field update commands.
package update

import (
	"errors"
	"fmt"
	"strings"
)

// FieldPath is a non-empty sequence of non-empty field names.
type FieldPath []string

// Update describes a change to one field.
type Update interface {
	FieldName() string
	FieldPath() FieldPath
	Value() any
}

func ByFieldName(fieldName string, value any) Update {
	if fieldName == "" {
		panic("fieldName cannot be empty")
	}
	if strings.Contains(fieldName, ".") {
		v := update{fieldPath: strings.Split(fieldName, "."), value: value}
		if err := v.Validate(); err != nil {
			panic(err)
		}
		return v
	}
	return update{fieldName: fieldName, value: value}
}

func ByFieldPath(fieldPath FieldPath, value any) Update {
	if len(fieldPath) == 0 {
		panic("fieldPath cannot be empty")
	}
	v := update{fieldPath: fieldPath, value: value}
	if err := v.Validate(); err != nil {
		panic(err)
	}
	return v
}

func DeleteByFieldName(fieldName string) Update {
	if fieldName == "" {
		panic("fieldName cannot be empty")
	}
	return update{fieldName: fieldName, value: DeleteField}
}

func DeleteByFieldPath(path ...string) Update {
	if len(path) == 0 {
		panic("fieldPath cannot be empty")
	}
	return update{fieldPath: path, value: DeleteField}
}

type update struct {
	fieldName string
	fieldPath FieldPath
	value     any
}

func (v update) FieldName() string { return v.fieldName }
func (v update) FieldPath() FieldPath { return v.fieldPath }
func (v update) Value() any { return v.value }

func (v update) Validate() error {
	if strings.TrimSpace(v.fieldName) == "" && len(v.fieldPath) == 0 {
		return errors.New("either fieldName or fieldPath must be provided")
	}
	if strings.Contains(v.fieldName, ".") {
		return fmt.Errorf("fieldName contains '.' character: %q", v.fieldName)
	}
	if v.fieldName != "" && len(v.fieldPath) > 0 {
		return fmt.Errorf("both FieldVal and fieldPath are provided: %v, %+v", v.fieldName, v.fieldPath)
	}
	for i, fp := range v.fieldPath {
		if strings.TrimSpace(fp) == "" {
			return fmt.Errorf("empty field path component at index %d", i)
		}
	}
	return nil
}

type sentinel int

const (
	DeleteField sentinel = iota
	ServerTimestamp
)
