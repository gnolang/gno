package main

func main() {
	assert()
	assert2()
}

func assert() {
	tests := []func(uint64, uint64) uint64{
		func(a, b uint64) uint64 {
			if b > 0 {
				return 1
			}
			return 0
		},
	}

	for _, test := range tests {
		shouldPanic := func(f func()) {
			f()
		}

		shouldPanic(func() { test(0, 1) })
	}
}

func assert2() {
	tests := []func(uint64, uint64) uint64{
		func(a, b uint64) uint64 {
			if b > 0 {
				return 1
			}
			return 0
		},
	}

	for _, test1 := range tests {
		shouldPanic := func(f func()) {
			f()
		}

		shouldPanic(func() { test1(0, 1) })
	}
}
