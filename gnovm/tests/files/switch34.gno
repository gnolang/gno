package main

func main() {
	var a any
	switch a.(type) {
	case []int:
	case []string:
	}
	println("bye")
}

// Output:
// bye
