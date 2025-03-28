package main

import (
	"bytes"
	"io"
)

func testImpl(v any) {
	defer func() {
		if r := recover(); r != nil {
			println(r)
		}
	}()

	_ = v.(io.Reader)
	println("ok")
}

func main() {
	{
		var v any
		var r io.Reader
		r = bytes.NewBuffer([]byte("hello"))
		v = r
		testImpl(v)
	}
	{
		var v any
		var r io.Reader
		v = r
		testImpl(v)
	}
	{
		var v any
		v = bytes.NewBuffer([]byte("hello"))
		testImpl(v)
	}
}

// Output:
// ok
// interface conversion: interface is nil, not io.Reader
// ok
