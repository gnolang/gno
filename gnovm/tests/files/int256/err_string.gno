package main

func main() {
	println(int256("1"))
}

// Error:
// main/err_string.gno:4:10-21: cannot convert StringKind to Int256Kind

// TypeCheckError:
// main/err_string.gno:4:17: cannot convert "1" (untyped string constant) to type int256
