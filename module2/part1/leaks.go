package leaks

import (
	"fmt"
	"runtime"
	"time"
)

type LeakDetector struct {
	numGBefore int
	numGAfter  int
}

func NewLeakDetector() *LeakDetector {
	return &LeakDetector{
		numGBefore: 0,
		numGAfter:  0,
	}
}

// Snapshot возвращает текущее число горутин.
func (d *LeakDetector) Snapshot() int {
	return runtime.NumGoroutine()
}

// AssertNoLeaks: снимок → f() → GC → пауза → снимок → сравнение.
// Возвращает ошибку с diff, если горутин стало больше.
func (d *LeakDetector) AssertNoLeaks(f func()) error {
	time.Sleep(10 * time.Millisecond)
	runtime.GC()
	d.numGBefore = d.Snapshot()
	f()
	runtime.GC()
	// Пауза нужна,чтобы все горутины завершили свою работу и на сбор
	time.Sleep(50 * time.Millisecond)
	d.numGAfter = d.Snapshot()

	if d.numGAfter-d.numGBefore > 0 {
		return fmt.Errorf("произошла утечка горутин до %d, после %d, разница %d", d.numGBefore, d.numGAfter, d.numGAfter-d.numGBefore)
	}

	return nil
}
