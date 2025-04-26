package main

func main() {
	defer func() {
		print(recover())
	}()
	defer func() {
		defer print(recover())
		defer panic("1")
		recover()
	}()
	defer recover()
	panic("2")
}

// Output:
// 21
