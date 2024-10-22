package main

func main() {
	rangeMap()
	rangeSlice()
}

func rangeMap() {
	m := map[int]string{999: "asdf"}
	var v int
	for v = range m {
		println(v)
	}
}

func rangeSlice() {
	m := []int{9, 4, 6}
	var v int
	for v = range m {
		println(v)
	}
}

// Output:
// 999
// 0
// 1
// 2
