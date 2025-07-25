package main

type String string

func main() {
	x := "STRING"
	y := String(x)
	println(x + y)
}

// Error:
// main/type31.gno:8:10-15: invalid operation: x<VPBlock(1,0)> + y<VPBlock(1,1)> (mismatched types string and main.String)

// TypeCheckError:
// main/type31.gno:8:10: invalid operation: x + y (mismatched types string and String)
