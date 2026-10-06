// False-positive guard for the method-expression visibility check at
// preprocess.go's *TypeType branch: same-package access to an unexported
// method via method-expression syntax must continue to work.

package main

type t struct {
	n int
}

func (x *t) inc() {
	x.n++
}

func main() {
	x := &t{n: 1}
	(*t).inc(x)
	println(x.n)
}

// Output:
// 2
