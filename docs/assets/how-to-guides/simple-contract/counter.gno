package counter

import "fmt"

var count int

func Increment() {
	count++
}

func Decrement() {
	count--
}

func Render(_ string) string {
	return fmt.Sprintf("Count: %d", count)
}
