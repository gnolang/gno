// MAXALLOC: 500000000
package main

// Test that switch statement case matching charges gas
// for string and array comparisons via doOpSwitchClauseCase.

func main() {
	buf := make([]byte, 1000)
	s := string(buf)

	// String switch: each case comparison should charge per-byte gas.
	switch s {
	case "a":
	case "b":
	case string(buf):
	}

	// Array switch: each case comparison should charge per-element gas.
	a := [5]int{1, 2, 3, 4, 5}
	switch a {
	case [5]int{0, 0, 0, 0, 0}:
	case [5]int{1, 2, 3, 4, 5}:
	}
}

// Gas:
// 175770
