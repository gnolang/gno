package main

import "fmt"

func theError() {
	for i := 0; i < 10; i++ {
		switch {
		case i%2 == 0:
			fmt.Println(i, "even")
		default:
			goto byebye
		}
		continue
	byebye:
		fmt.Println("does break")
		break
	}
}

func main() {
	theError()
}

// Output:
// 0 even
// does break
