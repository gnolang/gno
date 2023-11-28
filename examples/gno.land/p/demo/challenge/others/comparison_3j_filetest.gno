package main

func main() {
	var a int = 1
	a -= 1
	println(a) //0

	a += 1
	a *= 2
	println(a) //2

	a /= 2
	println(a) //1

	a += 2 //3
	a %= 2 //1
	println(a)
}

// Output:
// 0
// 2
// 1
// 1
