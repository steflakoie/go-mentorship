package main

import (
	"context"
	"fmt"
)

// Напиши функцию SafeClose(ch chan int),
// которая безопасно закрывает канал.
// В комментарии — почему двойной close паникует и как это можно (и нельзя) обойти.

func SafeClose(ch chan int) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Предотвращение повторного закрытия канала")
		}
	}()

	close(ch)
}

func SafeCloseContext(ctx context.Context, ch chan int) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Предотвращение повторного закрытия канала")
		}
	}()

	select {
	case v, ok := <-ch:
		if !ok {
			return
		}
		fmt.Println("еще идут значения", v)
	case <-ctx.Done():
		return
	}
}

func main() {

}
