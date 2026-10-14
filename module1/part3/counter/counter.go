package counter

type Valuer interface{ Value() int }

type ValueCounter struct {
	n int
}

type PtrCounter struct {
	n int
}

func (vc ValueCounter) Value() int {
	return vc.n
}

func (vc ValueCounter) WithIncrement() ValueCounter {
	return ValueCounter{n: vc.n + 1}
}

func (pc *PtrCounter) Value() int {
	return pc.n
}

func (pc *PtrCounter) Increment() {
	pc.n++
}

var _ Valuer = ValueCounter{}

// var _ Valuer = PtrCounter{} // потому что метод Value реализует *PtrCounter, но не он сам
var _ Valuer = (*PtrCounter)(nil)
