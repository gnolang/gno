package main

func main() {
	a := struct {
		A int `a`
	}{1}
	b := struct {
		A int `b`
	}{2}
	s := append([]interface{}{}, a, b)
	println(len(s))
}

// Output:
// 2
