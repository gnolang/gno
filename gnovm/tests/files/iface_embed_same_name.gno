package main

import "filetests/extern/ifaceext"

// omarsy's review case on PR #5739: I holds two distinct unexported methods
// both named sec (main.sec and the flattened ifaceext.sec). First-match-wins
// name lookup makes I fail its own implements check — even assigning a value
// of type I to type I panics "main.I does not implement main.I".

type I interface {
	sec() int
	ifaceext.Sec
}

func f(x I) int {
	var y I = x // assign to the SAME type
	return y.sec()
}

// T satisfies both: own main.sec plus promoted ifaceext.sec.
type T struct{ ifaceext.SecImpl }

func (T) sec() int { return 1 }

func main() {
	var t T
	println(f(t))
}

// Output:
// 1
