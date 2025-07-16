package main

func main() {
	const a = len(map[string][]int{"arr": {1, 2}})
	println("ok", a)
}

// Error:
// main/const48.gno:4:8-48: (const-type map[string][]int){(const ("arr" string)): (const-type []int){(const (1 int)), (const (2 int))}} (variable of type map[string][]int) is not constant

// TypeCheckError:
// main/const48.gno:4:12: len(map[string][]int{…}) (value of type int) is not constant
