package main

// Copying an array of interfaces shares the held values (ArrayValue.Copy).
// Every way of extracting a held value and mutating the result must leave
// both copies untouched; the output matches go run.

type T struct{ G int }

type S struct {
	F  int
	A  [2]int
	In T
	Ar [1]any
}

func (s S) Mut() int { s.F = 100; s.A[0] = 100; s.In.G = 100; return s.F }

type I interface{ Mut() int }

func ret(a [2]any) S { return a[0].(S) }

func main() {
	var a [2]any
	a[0] = S{F: 1, Ar: [1]any{S{F: 7}}}
	a[1] = [2]S{{F: 1}, {F: 1}}
	b := a

	// range over held array, value bound
	for _, v := range b[1].([2]S) {
		v.F = 9
		v.A[0] = 9
	}
	// method value on held value
	f := b[0].(S).Mut
	f()
	// interface method dispatch
	b[0].(I).Mut()
	var ii I = b[0].(S)
	ii.Mut()
	// returned without copy then used as receiver
	ret(b).Mut()
	// nested interface array held inside held struct
	c := b[0].(S).Ar
	c[0] = S{F: 8}
	// type switch binding
	switch v := b[1].(type) {
	case [2]S:
		v[0].F = 9
	}
	// define then mutate
	d := b[0].(S)
	d.F = 9
	d.Ar[0] = S{F: 9}
	// multi-assign from func
	e, _ := ret(b), 0
	e.F = 9
	// closure return
	g := func() S { return b[0].(S) }()
	g.In.G = 9

	println(a[0].(S).F, a[0].(S).A[0], a[0].(S).In.G, a[0].(S).Ar[0].(S).F, a[1].([2]S)[0].F, a[1].([2]S)[0].A[0])
	println(b[0].(S).F, b[0].(S).A[0], b[0].(S).In.G, b[0].(S).Ar[0].(S).F, b[1].([2]S)[0].F, b[1].([2]S)[0].A[0])
}

// Output:
// 1 0 0 7 1 0
// 1 0 0 7 1 0
