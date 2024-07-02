package main

import (
	"fmt"
	"io"
	"strings"
)

func main() {
	var mr io.Reader
	var buf []byte
	nread := 0
	withFooBar := func(tests func()) {
		r1 := strings.NewReader("foo ")
		//r2 := strings.NewReader("")
		//r3 := strings.NewReader("bar")
		//mr = io.MultiReader(r1, r2, r3)
		mr = io.MultiReader(r1)
		buf = make([]byte, 20)
		tests()
	}
	expectRead := func(size int, expected string, eerr error) {
		nread++
		n, gerr := mr.Read(buf[0:size])
		if n != len(expected) {
			panic(fmt.Sprintf("#%d, expected %d bytes; got %d",
				nread, len(expected), n))
		}
		got := string(buf[0:n])
		if got != expected {
			panic(fmt.Sprintf("#%d, expected %q; got %q",
				nread, expected, got))
		}
		if gerr != eerr {
			panic(fmt.Sprintf("#%d, expected error %v; got %v",
				nread, eerr, gerr))
		}
		buf = buf[n:]
	}
	withFooBar(func() {
		expectRead(5, "foo ", nil)
	})
	println("ok")
}

// Output:
// ok
