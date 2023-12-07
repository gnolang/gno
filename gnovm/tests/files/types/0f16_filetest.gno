package main

type word []int

var a [2]word
var c [2]word

func gen() word {
	return []int{1}
}

// TODO: should stop at comparable check
func main() {
	a = [2]word{gen(), gen()}
	c = [2]word{gen(), gen()}
	println(a == c)
}

// Error:
// main/files/types/0f16_filetest.gno:16: [2]main.word cannot be compared
