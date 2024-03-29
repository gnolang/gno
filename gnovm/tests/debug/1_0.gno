package main

import "fmt"

func main() {
	counter0 := 0
	counter1 := 0

	y := 0

	var fs []func()

	defer func() {
		for _, ff := range fs {
			ff()
		}
	}()

LOOP_START:
	if counter0 < 2 {
		counter1 = 0
		fmt.Printf("Outer loop start: counter0=%d\n", counter0)

	NESTED_LOOP_START:
		if counter1 < 2 {
			fmt.Printf("  Nested loop: counter1=%d\n", counter1)
			counter1++
			goto NESTED_LOOP_START
		}

		x := y
		fs = append(fs, func() { println(x) })

		fmt.Println("Exiting nested loop")
		counter0++
		y++
		goto LOOP_START
	} else {
		return
	}
}

// Output:
// Outer loop start: counter0=0
//   Nested loop: counter1=0
//   Nested loop: counter1=1
// Exiting nested loop
// Outer loop start: counter0=1
//   Nested loop: counter1=0
//   Nested loop: counter1=1
// Exiting nested loop
// 0
// 1
