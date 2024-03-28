package main

func main() {
	var buf []byte

	// when eval this, buf is still nil,
	expectRead := func(size int, expected string, err error) {
		b := buf[0:size]
		println(b)
		println(len(buf))
		println(cap(buf))
	}

	// buf should be captured here, here it's volatile, not where it defined
	// namely, the vp should point here, ->  so mutate offset
	withFooBar := func() {
		buf = make([]byte, 20)
		expectRead(5, "foo ", nil)
	}
	withFooBar()
}

// Output:
// slice[0x0000000000]
// 20
// 20
