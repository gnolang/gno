package main

func main() {
	t := make(map[int]string, -1)
	println(t)
}

// TypeCheckError:
// main/make7.gno:4:28: invalid argument: index -1 (constant of type int) must not be negative

// Error:
// main/make7.gno:4:7-31: invalid argument: index (-1 <untyped> bigint) must not be negative
