package main

// Struct composite literal: field names (i, v) match range loop var names.
// Field names must NOT be renamed to i.loopvar / v.loopvar.
// Value expressions (i, v) must be renamed correctly.
type S struct {
	i int
	v string
}

func main() {
	var fns []func() S
	for i, v := range []string{"a", "b", "c"} {
		i := i
		v := v
		fns = append(fns, func() S { return S{i: i, v: v} })
	}
	for _, fn := range fns {
		s := fn()
		println(s.i, s.v)
	}
}

// Output:
// 0 a
// 1 b
// 2 c
