package main

// print writes its operands with no separator; println separates them with
// spaces.
func main() {
	print("abc", "xyz", "\n")
	print(1, 2, 3, "\n")
	print(1, " ", true, "\n")
	println("abc", "xyz")
	println(1, 2, 3)
}

// Output:
// abcxyz
// 123
// 1 true
// abc xyz
// 1 2 3
