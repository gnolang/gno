package main

import "fmt"

func theErrorButAnotherOne() {
	if false {
	} else {
		for x := 0; x < 10; x++ {
			fmt.Println(x)
			if x == 1 {
				goto Match
			}
			continue
		}
	}
Match:
	println("match")
}

func main() {
	theErrorButAnotherOne()
}

// Output:
// 0
// 1
// match
