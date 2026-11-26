// Package t76lib holds the callees a realm aliases in a local variable.
package t76lib

// Each and Keep only call f: a literal passed in place is lent to them.
func Each(xs []int, f func(int)) {
	for _, x := range xs {
		f(x)
	}
}

func Keep(xs []int, f func(int)) {
	for i := len(xs) - 1; i >= 0; i-- {
		f(xs[i])
	}
}

// Call only calls f.
func Call(f func()) { f() }

// EachVar is a package variable holding a function, not a declaration.
var EachVar = Each

type Inner struct{}

func (Inner) Do(xs []int, f func(int)) { Each(xs, f) }
