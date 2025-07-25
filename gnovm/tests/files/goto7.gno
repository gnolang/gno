package main

import "fmt"

func main() {
	counter := 0
	for {
	LABEL:
		if true {
			if counter == 3 {
				break
			}
			counter++
			fmt.Println("true")
			goto LABEL
		}
	}
}

// Output:
// true
// true
// true
