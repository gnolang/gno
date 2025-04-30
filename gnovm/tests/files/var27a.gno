package main

var other = func() {
	if true {
		something := 2
		print(something) // 2
	} else {
		print(something) // a string, but single shared 'st' masks the outer/global reference.
	}
}
var something = "a string"

func main() {
	other()
}

// Output:
// 2
