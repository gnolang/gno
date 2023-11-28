// MAXALLOC: 100000000
// max total allocation of 100 mb.
package main

import "regexp"

var reName = regexp.MustCompile(`^[a-z]+[_a-z0-9]*$`)

func main() {
	for i := 0; i < 100; i++ {
		if !(reName.MatchString("thisisatestname")) {
			panic("error")
		}
	}
	println(true)
}

// Output:
// true
