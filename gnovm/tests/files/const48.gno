package main

func main() {
	const a = len(map[string][]int{"arr": {1, 2}})
	println("ok", a)
}

// Error:
// main/files/const48.gno:4:8: map[(const-type string)] [](const-type int){(const ("arr" string)): (const-type []int){(const (1 int)), (const (2 int))}} (variable of type map[string][]int) is not constant
