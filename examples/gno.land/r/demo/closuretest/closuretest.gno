package closuretest

import "strconv"

// The closures made in init capture a local and change package state. A
// stored closure writes only if it is crossing, so each takes cur realm and
// is called with cross.
var (
	count   int
	stepper func(cur realm) int
)

var (
	accumulator func(cur realm, val int)
	history     []int
)

func init() {
	step := 3
	stepper = func(cur realm) int {
		count += step
		return count
	}

	maxLen := 10
	history = make([]int, 0, maxLen)
	accumulator = func(cur realm, val int) {
		if len(history) < maxLen {
			history = append(history, val)
		}
	}
}

func Step(cur realm) string {
	result := stepper(cross(cur))
	return "count=" + strconv.Itoa(result)
}

func Accumulate(cur realm, val int) string {
	accumulator(cross(cur), val)
	return "history length=" + strconv.Itoa(len(history))
}

func Render(_ string) string {
	return "closuretest: count=" + strconv.Itoa(count) + " history=" + strconv.Itoa(len(history))
}
