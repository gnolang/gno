package main

func main() {
	var errors []error
	errors = append(errors, nil, nil)
	println(len(errors))
}

// Output:
// 2
