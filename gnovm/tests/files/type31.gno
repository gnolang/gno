package main

type String string

func main() {
	x := "STRING"
	y := String(x)
	println(x + y)
}

// Error:
// main/files/type31.gno:8:31: invalid operation: mismatched types string and main.String
