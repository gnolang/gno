package main

func main() {
	var buf []byte

	// when eval this, buf is still nil,
	expectRead := func(size int, expected string, eerr error) {
		b := buf[0:size]
		println(b)
	}

	// buf should be captured here, here it's volatile, not where it defined
	// namely, the vp should point here, ->  so mutate offset
	withFooBar := func() {
		buf = make([]byte, 20)
		expectRead(5, "foo ", nil)
	}
	withFooBar()
}
