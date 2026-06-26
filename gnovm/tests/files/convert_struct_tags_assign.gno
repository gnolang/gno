package main

func main() {
	var x struct {
		A int `a`
	}
	y := (struct {
		A int `b`
	})(x)
	var z struct {
		A int `b`
	} = y
	println(z.A)
}

// Output:
// 0
