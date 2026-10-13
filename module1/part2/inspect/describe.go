package inspect

import (
	"fmt"
	"reflect"
)

func Describe(v any) string {
	vType := reflect.TypeOf(v)
	vValue := reflect.ValueOf(v)

	return fmt.Sprintf("T:%v V:%v", vType, vValue)
}
