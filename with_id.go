package record

import "fmt"

// WithID is a record envelope with a strongly typed ID.
type WithID[K comparable] struct {
	ID     K      `json:"id"`
	FullID string `json:"fullID,omitempty"`
	Key    *Key   `json:"-"`
	Record Record `json:"-"`
}

func (v WithID[K]) String() string {
	if v.FullID == "" {
		return fmt.Sprintf("{ID=%v, FullID=nil, Key=%v, Record=%v}", v.ID, v.Key, v.Record)
	}
	if id, ok := any(v.ID).(string); ok {
		return fmt.Sprintf(`{ID="%s", FullID="%s", Key=%v, Record=%v}`, id, v.FullID, v.Key, v.Record)
	}
	return fmt.Sprintf(`{ID=%+v, FullID="%s", Key=%v, Record=%v}`, v.ID, v.FullID, v.Key, v.Record)
}

// NewWithID creates a typed envelope around key and data.
func NewWithID[K comparable](id K, key *Key, data any) WithID[K] {
	return WithID[K]{ID: id, Key: key, Record: NewRecordWithData(key, data)}
}
