package main

import (
	"bytes"
	"io"
)

func main() {
	{
		var v interface{}
		var r io.Reader
		r = bytes.NewBuffer([]byte("hello"))
		v = r
		if _, ok := v.(io.Reader); ok {
			println("ok")
		} else {
			println("not ok")
		}
	}
	{
		var v interface{}
		var r io.Reader
		v = r
		if _, ok := v.(io.Reader); ok {
			println("ok")
		} else {
			println("not ok")
		}
	}
	{
		var v interface{}
		v = bytes.NewBuffer([]byte("hello"))
		if _, ok := v.(io.Reader); ok {
			println("ok")
		} else {
			println("not ok")
		}
	}
}

// Output:
// ok
// not ok
// ok
