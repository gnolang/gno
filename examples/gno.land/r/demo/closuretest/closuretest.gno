package closuretest

import "strconv"

var (
	count   int
	stepper func() int
)

var (
	accumulator func(int)
	history     []int
)

func init() {
	step := 3
	stepper = func() int {
		count += step
		return count
	}

	history = make([]int, 0, 10)
	accumulator = func(val int) {
		history = append(history, val)
	}
}

func Step() string {
	result := stepper()
	return "count=" + strconv.Itoa(result)
}

func Accumulate(val int) string {
	accumulator(val)
	return "history length=" + strconv.Itoa(len(history))
}

func Render(_ string) string {
	return "closuretest: count=" + strconv.Itoa(count) + " history=" + strconv.Itoa(len(history))
}
