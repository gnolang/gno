package gnolang

import (
	"testing"

	"github.com/gnolang/overflow"
)

func BenchmarkOpAdd(b *testing.B) {
	m := NewMachine("bench", nil)
	x := TypedValue{T: IntType}
	x.SetInt(4)
	y := TypedValue{T: IntType}
	y.SetInt(3)

	b.ResetTimer()

	for range b.N {
		m.PushOp(OpHalt)
		m.PushExpr(&BinaryExpr{})
		m.PushValue(x)
		m.PushValue(y)
		m.PushOp(OpAdd)
		m.Run()
	}
}

//go:noinline
func AddNoOverflow(x, y int) int { return x + y }

func BenchmarkAddNoOverflow(b *testing.B) {
	x, y := 4, 3
	c := 0
	for range b.N {
		c = AddNoOverflow(x, y)
	}
	if c != 7 {
		b.Error("invalid result")
	}
}

func BenchmarkAddOverflow(b *testing.B) {
	x, y := 4, 3
	c := 0
	for range b.N {
		c = overflow.Addp(x, y)
	}
	if c != 7 {
		b.Error("invalid result")
	}
}

func TestOpAdd1(t *testing.T) {
	m := NewMachine("test", nil)
	a := TypedValue{T: IntType}
	a.SetInt(4)
	b := TypedValue{T: IntType}
	b.SetInt(3)
	t.Log("a:", a, "b:", b)

	start := m.NumValues
	m.PushOp(OpHalt)
	m.PushExpr(&BinaryExpr{})
	m.PushValue(a)
	m.PushValue(b)
	m.PushOp(OpAdd)
	m.Run()
	res := m.ReapValues(start)
	t.Log("res:", res)
}
