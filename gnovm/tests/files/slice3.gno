package main

func main() {
	i := []string{nil}
	println(i)
}

// Error:
// main/files/slice3.gno:4:7: cannot use nil as string value in array, slice literal or map literal

// TypeCheckError:
// main/files/slice3.gno:4:16: cannot use nil as string value in array or slice literal
