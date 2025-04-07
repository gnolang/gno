package main

func main() {
	defer func() {
		println("3")
	}()
	println("1")
	defer func() {
		println("2")
	}()
}

// Output:
// 1
// 2
// 3
