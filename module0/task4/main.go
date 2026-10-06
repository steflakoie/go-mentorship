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
	users := []*User{
		{ID: 10, Name: "Alice"},
		{ID: 42, Name: "Bob"},
	}

	var err error = &NotFoundError{ResourceID: 42}
	err = fmt.Errorf("repository error: %w", err)
	deepWrappedErr := fmt.Errorf("api handler failed: %w", err)

	if target, found := Task4(users, deepWrappedErr); found {
		fmt.Printf("[Успех] errors.As прошел по всей цепочке и достал ID: %d\n", target.ResourceID)
		fmt.Printf("Элемент с ID 42 в слайсе теперь: %v\n", users[1])
	}
}
