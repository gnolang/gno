package main

var number = 0

func SetNumber(n int) {
	number = n
}

func reset() {
	number = 0
}

func main() {
	defer reset()
	defer println(number)

	println(number)
	SetNumber(42)
	println(number)
}

// Output:
// 0
// 42
// 0
