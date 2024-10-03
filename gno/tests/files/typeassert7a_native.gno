package main

import (
	"bytes"
	"io"
)

func testImpl(v interface{}) {
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
		var v interface{}
		var r io.Reader
		r = bytes.NewBuffer([]byte("hello"))
		v = r
		testImpl(v)
	}
	{
		var v interface{}
		var r io.Reader
		v = r
		testImpl(v)
	}
	{
		var v interface{}
		v = bytes.NewBuffer([]byte("hello"))
		testImpl(v)
	}
}

// Output:
// ok
// interface conversion: interface is nil, not gonative{io.Reader}
// ok
