package inspect_test

import (
	"fmt"
	"go-mentorship/module1/part2/inspect"
	"testing"
)

// Ответ на вопрос
// err имеет type = *someError, а nil это nil/ У нас разные типы значений вот и поулчем nil

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
			res: "T:<nil> V:<nil>",
		},
		{
			v:   42,
			res: "T:int V:42",
		},
		{
			v:   "hi",
			res: "T:string V:hi",
		},
		{
			v:   err,
			res: "T:*inspect_test.someError V:<nil>",
		},
		{
			v:   realErr,
			res: "T:*inspect_test.someError V:err",
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
