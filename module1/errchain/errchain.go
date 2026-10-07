package errchain

import (
	"reflect"
)

// Unwrapper — интерфейс, который умеет отдавать следующую ошибку в цепочке.
type Unwrapper interface {
	Unwrap() error
}

func Unwrap(err error) error {
	u, ok := err.(Unwrapper)
	if !ok {
		return nil
	}

	return u.Unwrap()
}

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
	return false
}

func Errchain() {

}
