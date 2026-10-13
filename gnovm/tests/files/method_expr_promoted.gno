package main

type T0 struct{ n int }

func (t T0) v0() int           { return t.n }
func (t *T0) p0() int          { t.n++; return t.n }
func (t T0) sum(xs ...int) int { s := t.n; for _, x := range xs { s += x }; return s }

type T1 struct{ m int }

func (t *T1) p1() (int, string) { return t.m, "p1" }

type T3 struct {
	T0
	*T1
}

func main() {
	v := T3{T0{1}, &T1{5}}
	f := T3.v0
	println(f(v))
	g := (*T3).p0
	println(g(&v), v.n)
	h := T3.sum
	println(h(v, 1, 2, 3))
	k := T3.p1 // promoted through *T1: no pointer to T3 needed
	a, b := k(v)
	println(a, b)
	pv := (*T0).v0 // value method through a pointer: func(*T0) int
	println(pv(&v.T0))
}

// Output:
// 1
// 2 2
// 8
// 5 p1
// 2
