package leaks_test

import (
	leaks "go-mentorship/module2/part1"
	"sync"
	"testing"
	"time"
)

func TestNoLeaks_Normal(t *testing.T) {
	testF := func() {}

	d := leaks.NewLeakDetector()
	err := d.AssertNoLeaks(testF)
	if err != nil {
		t.Fatalf("unexpected leak %v", err)
	}
}

func TestNoLeaks_ProperGoroutine(t *testing.T) {
	testF := func() {
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
		}()
		wg.Wait()
	}

	d := leaks.NewLeakDetector()

	err := d.AssertNoLeaks(testF)
	if err != nil {
		t.Fatalf("unexpected leak %v", err)
	}
}

// main g закончила работу, висит g в тесте
func TestLeakDetected(t *testing.T) {
	done := make(chan struct{})
	testF := func() {
		go func() {
			select {
			case <-time.After(time.Hour):
			case <-done:
			}
		}()
	}
	t.Cleanup(func() {
		close(done)
		time.Sleep(10 * time.Millisecond) // дать горутине завершиться
	})

	d := leaks.NewLeakDetector()

	err := d.AssertNoLeaks(testF)
	if err == nil {
		t.Fatal("expected leak error, got nil")
	}
}

func TestLeakDetected_Channel(t *testing.T) {
	done := make(chan struct{})
	testF := func() {
		ch := make(chan int)
		go func() {
			select {
			case ch <- 1:
			case <-done:
			}
		}()
	}
	t.Cleanup(func() { close(done) })

	d := leaks.NewLeakDetector()

	err := d.AssertNoLeaks(testF)
	if err == nil {
		t.Fatal("expected leak error, got nil")
	}
}
