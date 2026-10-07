package errchain_test

import (
	"errors"
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
