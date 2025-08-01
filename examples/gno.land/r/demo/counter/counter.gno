package counter

import "strconv"

var counter int

func Increment(_ realm) int {
	counter++
	return counter
}

func Render(_ string) string {
	return strconv.Itoa(counter)
}
