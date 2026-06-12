package main

import "filetests/extern/aliasdef"

type Int = int

type B = struct{ Int }

type Stringer interface{ Str() string }
type SAlias = Stringer
type I interface{ SAlias }

type T struct{}

func (T) Str() string { return "ok" }

func main() {
	// a qualified alias embeds as its own name, so struct{aliasdef.Int}
	// is identical to struct{Int}
	var x, y interface{} = B{}, struct{ aliasdef.Int }{}
	if x != y {
		panic("FAIL: struct{Int} != struct{aliasdef.Int}")
	}

	// pointer embeds also use the alias name
	x, y = struct{ *Int }{}, struct{ *int }{}
	if x == y {
		panic("FAIL: struct{*Int} == struct{*int}")
	}

	// the embedded field is accessible by the alias name
	b := B{7}
	b.Int++
	println(b.Int)

	// embedding an alias in an interface still flattens the method set
	var v interface{} = T{}
	s, ok := v.(I)
	if !ok {
		panic("FAIL: T does not implement interface{SAlias}")
	}
	println(s.Str())
}

// Output:
// 8
// ok
