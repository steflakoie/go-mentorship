package main

import (
	"errors"
	"fmt"
)

// Работем с pointer receiver, так как меняем состояние стэка
// И в методе pop, произошла бы утечка из-за работы с копией

type Stack[T any] struct {
	elements []T
}

func (s *Stack[T]) Push(el T) {
	s.elements = append(s.elements, el)
}

func (s *Stack[T]) Pop() (T, error) {
	var zero T
	if s.IsEmpty() {
		return zero, errors.New("Stack is empty")
	}

	index := s.Len() - 1
	element := s.elements[index]

	// Если элемент стркутура или указатель
	s.elements[index] = zero

	s.elements = s.elements[:index]
	return element, nil
}

func (s *Stack[T]) Peek() (T, error) {
	if s.IsEmpty() {
		var zero T
		return zero, errors.New("Stack is empty")
	}

	return s.elements[s.Len()-1], nil
}

func (s *Stack[T]) Len() int {
	return len(s.elements)
}

func (s *Stack[T]) IsEmpty() bool {
	return len(s.elements) == 0
}

func main() {
	intStack := Stack[int]{}

	intStack.Push(10)
	intStack.Push(20)
	intStack.Push(30)

	fmt.Println("Размер стека:", intStack.Len()) // Выведет: 3

	top, _ := intStack.Peek()
	fmt.Println("Верхний элемент:", top) // Выведет: 30

	val, _ := intStack.Pop()
	fmt.Println("Извлечен элемент:", val) // Выведет: 30

	fmt.Println("Пуст ли стек?", intStack.IsEmpty()) // Выведет: false

	top, _ = intStack.Peek()
	fmt.Println("Верхний элемент:", top) // Выведет: 20
}
