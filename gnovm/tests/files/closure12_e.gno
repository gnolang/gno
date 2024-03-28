package main

type queueOnePass struct {
	sparse          []uint32
	dense           []uint32
	size, nextIndex uint32
}

func newQueue(size int) (q *queueOnePass) {
	return &queueOnePass{
		sparse: make([]uint32, size),
		dense:  make([]uint32, size),
	}
}
func main() {
	var (
		visitQueue = newQueue(10)
	)
	f := func() {
		println(visitQueue.size)
	}

	f()
}

// Output:
// 0
