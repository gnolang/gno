package main

import "fmt"

func theErrorButAnotherOne() {
	for x := 0; x < 10; x++ {
		fmt.Println(x)
		for y := 0; y < 10; y++ {
			goto byebye
		}
		continue
	byebye:
		break
	}
}

func main() {
	theErrorButAnotherOne()
}

// Output:
// 0
