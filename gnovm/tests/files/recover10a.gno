package main

func main() {
	defer func() {
		println(recover())
	}()

	panic("ahhhhh")
}

// Output:
// ahhhhh
