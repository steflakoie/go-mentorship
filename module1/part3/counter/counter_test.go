package counter

import (
	"testing"
)

// var valuer counter.Valuer = counter.PtrCounter{} // не скомпилируется
// Value() имеет pointer receiver, поэтому он находится в method set *PtrCounter,
// но не в method set PtrCounter. Значит PtrCounter не реализует Valuer, а *PtrCounter — реализует

func TestCounter(t *testing.T) {
	vc := ValueCounter{}
	pc := PtrCounter{}

	t.Run("vc test", func(t *testing.T) {
		vc.WithIncrement()
		if vc.Value() != 0 {
			t.Errorf("vc changed own Value = %v", vc.Value())
		}
	})

	t.Run("pc test", func(t *testing.T) {
		pc.Increment()
		if pc.Value() != 1 {
			t.Errorf("pc do not changed own Value = %v", pc.Value())
		}
	})

	t.Run("valuer test", func(t *testing.T) {
		vcNew := vc.WithIncrement()
		if vcNew.Value() != 1 {
			t.Errorf("pc do not changed own Value = %v", vcNew.Value())
		}
	})

	t.Run("key test", func(t *testing.T) {
		var v Valuer = ValueCounter{n: 10}
		if v.Value() != 10 {
			t.Errorf("v.n is not 10 v.n = %v", v.Value())
		}

		pc := &PtrCounter{n: 10}
		var v2 Valuer = pc
		pc.Increment()
		if v2.Value() != 11 {
			t.Errorf("v2.n is not 10 v2.n = %v", v2.Value())
		} // важно: v2 хранит указатель, изменения видны
	})
}
