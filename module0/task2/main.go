package main

import (
	"sync"
)

// двойной close паникует, так как это закрытие закрытого кннала.
// Обойти можно с помощью recover или sync.once.Do

type SafeChannel struct {
	ch   chan int
	once sync.Once
}

func (sc *SafeChannel) SafeClose() {
	sc.once.Do(func() {
		close(sc.ch)
	})
}
func SafeClose(ch chan int, once *sync.Once) {
	once.Do(func() {
		close(ch)
	})
}
func main() {

}
