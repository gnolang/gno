package main

func testEql(want, got interface{}) {
	if want != got {
		println(false)
	} else {
		println(true)
	}
}

// return var of error nil
func gen() error {
	return nil
}

func main() {
	r := gen()
	testEql(r, error(nil))
}

// Output:
// true
