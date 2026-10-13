package inspect

import (
	"fmt"
	"reflect"
)

var errorType = reflect.TypeOf((*error)(nil)).Elem()

func Describe(v any) string {
	// Даст true, только когда nil. Typed nil = false
	if v == nil {
		return "nil interface"
	}

	vValue := reflect.ValueOf(v)
	vType := vValue.Type()
	vKind := vValue.Kind()
	switch vKind {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		if vValue.IsNil() {
			return fmt.Sprintf("typed nil of %v", vType)
		}
	}
	return fmt.Sprintf("%v: %v", vType, vValue)
}
