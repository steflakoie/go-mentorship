package inspect_test

import (
	"fmt"
	"go-mentorship/module1/part2/inspect"
	"testing"
)

// Ответ на вопрос
// Typed nil под капотом содержит ненулевой указатель на тип(itab)
// Интерфейс равен nil, только когда оба слова пустые, и тип, и значение.

type someError struct {
	message string
}

func (e *someError) Error() string {
	return e.message
}

func TestDescribe(t *testing.T) {
	var e *someError = nil
	var err error = e
	var realErr error = &someError{message: "err"}
	testCases := []struct {
		v   any
		res string
	}{
		{
			v:   nil,
			res: "nil interface",
		},
		{
			v:   42,
			res: "int: 42",
		},
		{
			v:   "hi",
			res: "string: hi",
		},
		{
			v:   err,
			res: "typed nil of *inspect_test.someError",
		},
		{
			v:   realErr,
			res: "*inspect_test.someError: err",
		},
		{
			v:   []int{1, 2, 3},
			res: "[]int: [1 2 3]",
		},
		{
			v:   struct{ f []string }{f: []string{"a"}},
			res: "struct { f []string }: {[a]}",
		},
		{
			v:   struct{ f []string }{},
			res: "struct { f []string }: {[]}",
		},
		{
			v:   (chan int)(nil),
			res: "typed nil of chan int",
		},
		{
			v:   []int(nil),
			res: "typed nil of []int",
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Test #%v %v", i+1, tc.v), func(t *testing.T) {
			if funcRes := inspect.Describe(tc.v); funcRes != tc.res {
				t.Errorf("Describe: got %v, want: %v", funcRes, tc.res)
			}
		})
	}
}
