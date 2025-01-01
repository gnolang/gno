package main

func main() {
	const a = len(map[string][]int{"arr": {1, 2}})
	println("ok", a)
}

// Error:
// main/files/const48.gno:4:8: map[(const-type string)] [](const-type int) (variable of type map[string][]int) is not constant
