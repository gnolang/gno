// Test that a type switch distinguishes an interface with no dynamic
// type (enters `case nil`) from an interface holding a typed nil
// pointer like (*S)(nil), which has dynamic type *S and enters
// `case *S` with a nil underlying value.

package main

type S struct{}

func classify(x any) {
	switch v := x.(type) {
	case nil:
		println("nil", v == nil)
	case *S:
		println("*S", v == nil)
	default:
		panic("unexpected type")
	}
}

func main() {
	classify(nil)
	classify((*S)(nil))
	classify(&S{})
}

// Output:
// nil true
// *S true
// *S false
