package main

var errs = map[int]error{0: nil}

func main() {
	println(errs)
}

// Output:
// map{(0 int):(undefined)}
