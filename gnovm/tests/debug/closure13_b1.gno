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
	withFooBar := func() func() {
		buf = make([]byte, 20)
		return func() {
			b := buf[0:4]
			println(b)
		}
	}
	withFooBar()
	println("ok")
}

// Output:
// ok
