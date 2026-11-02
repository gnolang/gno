package main

func main() {
	var one uint256 = 1
	var s int256 = -1
	println(one << s)
}

// Error:
// runtime error: negative shift amount: (-1 int256)
