package main

import (
	"errors"
	"fmt"
)

type User struct {
	ID   int
	Name string
}

type NotFoundError struct {
	ResourceID int
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("User with id='%d' has not found.", e.ResourceID)
}

func Task4(users []*User, errAny any) (*NotFoundError, bool) {
	err, ok := errAny.(error)
	if !ok {
		return nil, false
	}

	var target *NotFoundError
	if errors.As(err, &target) {
		fmt.Printf("%v", target.ResourceID)
		return target, true
	}

	return nil, false
}

// error.Is не подойдет так как она просто проверяет совпадение ошибки,а error.As еще и извлекает данные
func main() {

}
