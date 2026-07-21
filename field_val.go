package record

import (
	"errors"
	"strings"
)

// FieldVal holds a named component of a composite key.
type FieldVal struct {
	Name  string `json:"NewFieldRef"`
	Value any    `json:"value"`
}

// Validate validates the field value.
func (v FieldVal) Validate() error {
	if strings.TrimSpace(v.Name) == "" {
		return errors.New("missing field name")
	}
	return nil
}
