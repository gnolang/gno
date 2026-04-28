// MAXALLOC: 500000000
package main

// Tests gas metering for switch-case matching on strings and arrays.
func main() {
	buf := make([]byte, 1000)
	s := string(buf)

	switch s {
	case "a":
	case "b":
	case string(buf):
	}

	a := [5]int{1, 2, 3, 4, 5}
	switch a {
	case [5]int{0, 0, 0, 0, 0}:
	case [5]int{1, 2, 3, 4, 5}:
	}
}

// Gas:
// 16824
