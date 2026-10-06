package main

import (
	"fmt"
)

// двойной close паникует, так как это закрытие закрытого кннала.
// Обойти можно с помощью recover или sync.once.Do

func SafeClose(ch chan int) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Предотвращение повторного закрытия канала")
		}
	}()

	close(ch)
}

func main() {

}
