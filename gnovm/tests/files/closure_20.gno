package main

import (
	"fmt"
)

func main() {
	i := 0
	var fs []func()

	defer func() {
		for _, ff := range fs {
			ff()
		}
	}()

	if i == 0 {
		goto Start
	}

Start:
	if i <= 5 {
		x := i
		fmt.Println("i: ", i)
		i++
		fs = append(fs, func() { println(x) })

		goto Start // Jumps back to the label "Start"
	}

	if i <= 10 {
		x := i
		fmt.Println("Counting past 5:", i)
		i++
		fs = append(fs, func() { println(x) })
		goto Start // Another jump back to the label "Start"
	}

	fmt.Println("Finished counting")
}

// Output:
// i:  0
// i:  1
// i:  2
// i:  3
// i:  4
// i:  5
// Counting past 5: 6
// Counting past 5: 7
// Counting past 5: 8
// Counting past 5: 9
// Counting past 5: 10
// Finished counting
// 0
// 1
// 2
// 3
// 4
// 5
// 6
// 7
// 8
// 9
// 10
