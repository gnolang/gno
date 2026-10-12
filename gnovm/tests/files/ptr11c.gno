// A nil-pointer value-method call via an interface panics with a value that
// implements the error interface (as in Go), so recover().(error) works.
// Mirrors Go's test/fixedbugs/issue19040.go.
package main

type T int

type I interface {
	F()
}

func (t T) F() {}

var (
	t *T
	i I = t
)

func main() {
	defer func() {
		println(recover().(error).Error())
	}()
	i.F()
}

// Output:
// value method main.T.F called using nil *T pointer
