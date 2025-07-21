package main

func f(x int) { return x }

func main() {
	print("hello")
}

// Error:
// main/fun23.gno:3:17-25: expected 0 return values; got 1

// TypeCheckError:
// main/fun23.gno:3:24: too many return values
// 	have (int)
// 	want ()
