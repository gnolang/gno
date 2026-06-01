package main

// Regression: `for _ = range slice/array` panicked
// "runtime error: invalid memory address or nil pointer dereference"
// from inside (*RangeStmt).AssertCompatible -> assertIndexTypeIsInt(nil)
// at gnolang/type_check.go. evalStaticTypeOf returns nil for a blank
// key; the SliceType/ArrayType branches called kt.Kind() without the
// nil guard the StringKind branch already had.
//
// Repro reduced from Go corpus fixedbugs/bug406.go.

type matrix struct {
	e []int
}

func (a matrix) equal() bool {
	for _ = range a.e {
	}
	for range a.e {
	}
	return true
}

func main() {
	var a matrix
	var i interface{}
	i = true && a.equal()
	_ = i
}

// Output:
