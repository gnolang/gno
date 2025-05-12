package main

var other = func() {
	if true {
		something := 2
		println(something) // 2
	}
	something += a
	println(something) // a string, but single shared 'st' masks the outer/global reference.
}
var something = "a string"
var a = "!"

func main() {
	other()
}

// Output:
// 2
// a string!
