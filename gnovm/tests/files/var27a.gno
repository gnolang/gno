package main

var other = func() {
	if true {
		something := 2
		print(something) // 2
	} else {
		anything := something
		println(anything)
	}
}

var something = "a string"

func main() {
	other()
}

// Output:
// 2
