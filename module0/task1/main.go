package main

import "fmt"

// Задача 1. Напиши функцию RemoveDuplicates([]int) []int, которая не мутирует исходный slice.
// Объясни, почему наивная реализация через append в цикле может «сломать» вызывающего.

// append(slice[:i], slice[i+1:]...)  ломает,
// он смещает элементы влево и но длина остается той же, добаляется лишний элемент

func RemoveDuplicates(slice []int) []int {
	nonDuplicates := map[int]int{}
	res := make([]int, 0, len(slice))

	for _, v := range slice {
		if _, ok := nonDuplicates[v]; !ok {
			nonDuplicates[v] = v
			res = append(res, v)
		}
	}

	return res
}

func main() {
	original := []int{1, 2, 2, 3, 4, 3, 5, 1}

	// Вызов функции
	cleaned := RemoveDuplicates(original)

	fmt.Println("Оригинальный слайс:", original) // [1 2 2 3 4 3 5 1] (не изменился)
	fmt.Println("Очищенный слайс:", cleaned)     // [1 2 3 4 5]
}
