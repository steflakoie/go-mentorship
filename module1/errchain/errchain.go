package errchain

import (
	"reflect"
)

// Is проверяет, есть ли target в цепочке err (по значению, через ==).
// Идёт по цепочке через Unwrapper.
func Is(err, target error) bool {
	if err == nil || target == nil {
		return err == target
	}
	// Если кастомная ошибка дает slice, map и т.д.
	// Сравнение невозможно
	isComparable := reflect.TypeOf(target).Comparable()

	if isComparable && err == target {
		return true
	}

	// Нужно если custom error реализует свой метод Is
	if x, ok := err.(interface{ Is(error) bool }); ok && x.Is(target) {
		return true
	}

	u, ok := err.(interface {
		Unwrap() error
	})
	if ok {
		err = u.Unwrap()
		return Is(err, target)
	}

	un, ok := err.(interface {
		Unwrap() []error
	})
	if ok {
		for _, err := range un.Unwrap() {
			if Is(err, target) {
				return true
			}
		}

		return false
	}

	return false
}

// As ищет в цепочке err ошибку, которую можно type-assert'нуть к типу target.
// target — указатель на переменную-интерфейс или указатель на конкретный тип.
// При успехе — присваивает и возвращает true.
func As(err error, target any) bool {
	if err == nil {
		return false
	}

	if target == nil {
		panic("errors: target cannot be nil")
	}

	targetVal := reflect.ValueOf(target)

	typ := targetVal.Type()
	if typ.Kind() != reflect.Pointer || targetVal.IsNil() {
		panic("errors: target must be a non-nil pointer")

	}

	targetType := typ.Elem()
	if targetType.Kind() != reflect.Interface && !targetType.Implements(errorType) {
		panic("errors: *target must be interface or implement error")
	}

	return as(err, target, targetType, targetVal)

}

func as(err error, target any, targetType reflect.Type, targetVal reflect.Value) bool {
	for err != nil {

		if reflect.TypeOf(err).AssignableTo(targetType) {
			targetVal.Elem().Set(reflect.ValueOf(err))

			return true
		}

		if x, ok := err.(interface{ As(any) bool }); ok && x.As(target) {
			return true
		}

		switch u := err.(type) {
		case interface{ Unwrap() error }:
			err = u.Unwrap()
		case interface{ Unwrap() []error }:
			for _, wErr := range u.Unwrap() {
				if as(wErr, target, targetType, targetVal) {
					return true
				}
			}

			return false
		default:
			return false
		}
	}

	return false
}

var errorType = reflect.TypeOf((*error)(nil)).Elem()

func Errchain() {

}
