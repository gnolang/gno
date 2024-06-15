package main

type String string

func main() {
	x := "STRING"
	y := String(x)
	println(x + y)
}

// Error:
// main/files/type31.gno:8:10: incompatible types in binary expression: string ADD main.String
