package counter

import "gno.land/p/demo/ufmt"

var counter int = 0

func Incr() {
	counter += 1
}

func Reset() {
	counter = 0
}

func Render() string {
	return ufmt.Sprintf("Counter: %d", counter)
}
