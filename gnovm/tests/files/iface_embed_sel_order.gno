package main

import "filetests/extern/ifaceext"

// Two distinct unexported methods named sec (main.sec and ifaceext.sec) in
// one interface. Selector resolution from main must find main's own sec
// regardless of declaration order (davd-gzl's review case on PR #5739).

// embed-first: foreign sealed ifaceext.sec listed before main's own sec.
func selEmbedFirst(x interface {
	ifaceext.Sec
	sec() int
}) int {
	return x.sec()
}

// own-first: main's own sec listed before the foreign embed.
func selOwnFirst(x interface {
	sec() int
	ifaceext.Sec
}) int {
	return x.sec()
}

// T satisfies both: own main.sec plus promoted ifaceext.sec.
type T struct{ ifaceext.SecImpl }

func (T) sec() int { return 1 }

func main() {
	var t T
	println(selEmbedFirst(t))
	println(selOwnFirst(t))
}

// Output:
// 1
// 1
