package main

type word []int

var a [2]word
var c [2]word

func gen() word {
	return []int{1}
}

// TODO: consider log desc
func main() {
	a = [2]word{gen(), gen()}
	c = [2]word{gen(), gen()}
	println(a == c)
}

// Error:
// main/eql_0f16.gno:16:10-16: [2]main.word is not comparable

// TypeCheckError:
// main/eql_0f16.gno:16:10: invalid operation: a == c ([2]word cannot be compared)
