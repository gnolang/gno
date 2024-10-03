package main

func rec() {
	println(recover())
}

func main() {
	defer rec()

	panic("ahhhhh")
}

// Output:
// ahhhhh
