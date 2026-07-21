package record

import (
	"fmt"
	"reflect"
)

// DataWithID is a typed record envelope with strongly typed data.
type DataWithID[K comparable, D any] struct {
	WithID[K]
	Data D
}

// NewDataWithID creates a typed data envelope. data must be a non-nil pointer
// or interface holding a struct or map.
func NewDataWithID[K comparable, D any](id K, key *Key, data D) DataWithID[K, D] {
	if key == nil {
		panic(fmt.Sprintf("key is nil for (id=%v)", id))
	}
	v := reflect.ValueOf(data)
	switch v.Kind() {
	case reflect.Pointer, reflect.Interface:
		if v.IsNil() {
			t := reflect.TypeOf(data)
			panic(fmt.Sprintf("data of type %v is nil for (id=%v, key=%v)", t.String(), id, key))
		}
		elemType := v.Elem().Type()
		switch elemType.Kind() {
		case reflect.Struct, reflect.Map:
		default:
			panic("data should be a pointer to a struct or map, got " + elemType.String())
		}
	default:
		t := reflect.TypeOf(data)
		if t == nil {
			panic(fmt.Sprintf("data is nil for (id=%v, key=%v)", id, key))
		}
		panic(fmt.Sprintf("data should be a pointer or an interface, got %v for (id=%v, key=%v)", t.String(), id, key))
	}
	return DataWithID[K, D]{WithID: NewWithID(id, key, data), Data: data}
}
