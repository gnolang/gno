// An embedded interface contributes only its method set to identity, never
// its own (alias) name: anonymous interfaces with the same flattened method
// set share one TypeID, matching Go. Covers embed-vs-explicit, alias-vs-
// target, and diamond embedding (the same method reached twice dedups).
package main

type Stringer interface{ Str() string }
type SAlias = Stringer

type R interface{ Read() int }
type W interface{ Write() int }
type A interface{ R }
type B interface{ R }

func main() {
	// embedding a named interface flattens to its methods, so an embed is
	// identical to spelling the method out
	var a, b interface{} = struct {
		X interface{ Stringer }
	}{}, struct {
		X interface{ Str() string }
	}{}
	if a != b {
		panic("FAIL: interface{Stringer} != interface{Str() string}")
	}

	// an embedded alias is identical to embedding its target
	var c, d interface{} = struct {
		X interface{ SAlias }
	}{}, struct {
		X interface{ Stringer }
	}{}
	if c != d {
		panic("FAIL: interface{SAlias} != interface{Stringer}")
	}

	// embedding two interfaces vs the union of their methods
	type RW = interface {
		Read() int
		Write() int
	}
	var e, f interface{} = struct {
		X interface {
			R
			W
		}
	}{}, struct{ X RW }{}
	if e != f {
		panic("FAIL: interface{R;W} != interface{Read;Write}")
	}

	// diamond: interface{A;B} reaches R's method twice; dedups to {Read},
	// identical to interface{R}
	var g, h interface{} = struct {
		X interface {
			A
			B
		}
	}{}, struct {
		X interface{ R }
	}{}
	if g != h {
		panic("FAIL: diamond interface{A;B} != interface{R}")
	}

	println("ok")
}

// Output:
// ok
