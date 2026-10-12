package errchain_test

import (
	"errors"
	"fmt"
	errchain "go-mentorship/module1/errchain"
	"testing"
)

type multiErr []error

func (m multiErr) Error() string   { return "multiError" }
func (m multiErr) Unwrap() []error { return []error(m) }

type errorUncomparable struct {
	f []string
}

func (errorUncomparable) Error() string {
	return "uncomparable error"
}

func (errorUncomparable) Is(target error) bool {
	_, ok := target.(errorUncomparable)
	return ok
}

type wrapped struct {
	msg string
	err error
}

func (e wrapped) Error() string { return e.msg }
func (e wrapped) Unwrap() error { return e.err }

func TestIs(t *testing.T) {
	err1 := errors.New("1")
	erra := wrapped{"wrap 2", err1}
	errb := wrapped{"wrap 3", erra}

	err3 := errors.New("3")

	testCases := []struct {
		err    error
		target error
		match  bool
	}{
		{nil, nil, true}, //0
		{nil, err1, false},
		{err1, nil, false},
		{err1, err1, true},
		{erra, err1, true},
		{errb, err1, true},
		{err1, err3, false},
		{erra, err3, false},
		{errb, err3, false},
		{errorUncomparable{}, errorUncomparable{}, true},    //9
		{errorUncomparable{}, &errorUncomparable{}, false},  //10
		{&errorUncomparable{}, errorUncomparable{}, true},   //11
		{&errorUncomparable{}, &errorUncomparable{}, false}, //12
		{errorUncomparable{}, err1, false},                  //13
		{&errorUncomparable{}, err1, false},
		{multiErr{}, err1, false},
		{multiErr{err1, err3}, err1, true},
		{multiErr{err3, err1}, err1, true},
		{multiErr{err1, err3}, errors.New("x"), false},
		{multiErr{err3, errb}, errb, true},
		{multiErr{err3, errb}, erra, true},
		{multiErr{err3, errb}, err1, true},
		{multiErr{errb, err3}, err1, true},
		{multiErr{nil}, nil, false},
	}

	for _, tt := range testCases {
		t.Run("", func(t *testing.T) {
			if got := errchain.Is(tt.err, tt.target); got != tt.match {
				t.Errorf("Is(%v, %v) = %v, want %v", tt.err, tt.target, got, tt.match)
			}
		})
	}
}

type CustomError struct {
	ID   int
	Text string
}

func (e CustomError) Error() string {
	return fmt.Sprintf("ошибка %d: %s", e.ID, e.Text)
}

// Тестовый интерфейс для проверки распаковки в интерфейсы
type Detailer interface {
	GetDetails() string
}

type DetailedError struct {
	Details string
}

func (e DetailedError) Error() string      { return e.Details }
func (e DetailedError) GetDetails() string { return e.Details }

// --- Тесты ---

func TestAs_Success(t *testing.T) {
	origErr := CustomError{ID: 42, Text: "база данных упала"}
	wrappedErr := fmt.Errorf("уровень 2: %w", fmt.Errorf("уровень 1: %w", origErr))

	var target CustomError

	// Проверяем, что функция возвращает true
	if !errchain.As(wrappedErr, &target) {
		t.Fatalf("As вернул false, хотя ошибка CustomError есть в цепочке")
	}

	// Проверяем, что данные успешно скопировались в target
	if target.ID != 42 || target.Text != "база данных упала" {
		t.Errorf("Данные внутри target искажены или не скопировались: %+v", target)
	}
}

func TestAs_NotFound(t *testing.T) {
	origErr := errors.New("простая стандартная ошибка")
	var target CustomError

	if errchain.As(origErr, &target) {
		t.Errorf("As вернул true, хотя CustomError в цепочке нет")
	}
}

func TestAs_InterfaceTarget(t *testing.T) {
	origErr := DetailedError{Details: "расширенное описание проблемы"}
	wrappedErr := fmt.Errorf("wrap: %w", origErr)

	// Целью выступает интерфейс Detailer, а не конкретная структура
	var target Detailer

	// ВАЖНО: Ваша старая реализация упадет здесь в панику (reflect.Set: value of type... не assignable)
	// Исправленная версия с .Convert() отработает успешно.
	if !errchain.As(wrappedErr, &target) {
		t.Fatalf("As вернул false при поиске по интерфейсу Detailer")
	}

	if target.GetDetails() != "расширенное описание проблемы" {
		t.Errorf("Получены неверные данные через интерфейс: %s", target.GetDetails())
	}
}

func TestAs_MultipleUnwrap(t *testing.T) {
	err1 := CustomError{ID: 100, Text: "первая"}
	err2 := errors.New("просто текст")

	// errors.Join объединяет ошибки через Unwrap() []error
	joinedErr := errors.Join(err2, err1)

	var target CustomError
	if !errchain.As(joinedErr, &target) {
		t.Errorf("As не смог найти ошибку внутри errors.Join (Unwrap() []error)")
	}
	if target.ID != 100 {
		t.Errorf("Неверный ID извлеченной ошибки: %d", target.ID)
	}
}

func TestAs_Panics(t *testing.T) {
	err := CustomError{ID: 1, Text: "тест"}

	// 1. Проверка паники на nil-таргет
	t.Run("nil target", func(t *testing.T) {
		defer func() { recover() }()
		errchain.As(err, nil)
		t.Errorf("Функция не выбросила панику при target = nil")
	})

	// 2. Проверка паники, если передан не указатель
	t.Run("not a pointer", func(t *testing.T) {
		defer func() { recover() }()
		var target CustomError
		errchain.As(err, target) // Передаем значение, а не &target
		t.Errorf("Функция не выбросила панику, когда target не является указателем")
	})

	// 3. Проверка паники, если указатель не на ошибку или не на интерфейс
	t.Run("invalid pointer type", func(t *testing.T) {
		defer func() { recover() }()
		var target int // Указатель на int не реализует error и не интерфейс
		errchain.As(err, &target)
		t.Errorf("Функция не выбросила панику, когда target указывает на int")
	})
}

type customAsError struct{}

func (e *customAsError) Error() string {
	return "custom error"
}

func (e *customAsError) As(target any) bool {
	p, ok := target.(**customAsTarget)
	if !ok {
		return false
	}

	*p = &customAsTarget{}
	return true
}

type customAsTarget struct{}

func (e *customAsTarget) Error() string {
	return "custom target"
}

func TestAs_CustomAsMethod(t *testing.T) {
	var target *customAsTarget

	err := error(&customAsError{})

	if !errchain.As(err, &target) {
		t.Fatal("As() = false, want true")
	}

	if target == nil {
		t.Fatal("target = nil, want non-nil")
	}
}

type multiError struct {
	errs []error
}

func (e *multiError) Error() string {
	return "multiple errors"
}

func (e *multiError) Unwrap() []error {
	return e.errs
}

type targetError struct{}

func (e *targetError) Error() string {
	return "target error"
}

func TestAs_NilInsideUnwrap(t *testing.T) {
	wrapped := &multiError{
		errs: []error{
			nil,
			&targetError{},
		},
	}

	var target *targetError

	if !errchain.As(wrapped, &target) {
		t.Fatal("As() = false, want true")
	}

	if target == nil {
		t.Fatal("target = nil, want non-nil")
	}
}
