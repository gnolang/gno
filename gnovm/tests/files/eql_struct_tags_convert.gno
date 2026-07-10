package main

func main() {
	x := struct {
		A int `a`
	}{1}
	y := (struct {
		A int `b`
	})(x)
	var i interface{} = y
	var j interface{} = struct {
		A int `b`
	}{1}
	println(i == j)
}

// Output:
// true
