// This is a sample target for the debugger's map reads.
package main

type S struct{ F int }

func main() {
	m := map[int]S{}
	println(m[1].F, len(m))
}
