package main

type S struct {
	i int
}

func main() {
	sArr := make([]S, 0, 4)
	sArr = append(sArr, S{1}, S{2}, S{3})
	println(sArr[0].i, sArr[1].i, sArr[2].i)

	newArr := append(sArr[:0], sArr[1:]...)

	// The append modified the underlying array because it was within capacity.
	println(len(sArr) == 3)
	println(sArr[0].i, sArr[1].i, sArr[2].i)

	// It generated a new slice that references the same array.
	println(len(newArr) == 2)
	println(newArr[0].i, newArr[1].i)

	// The struct should have been copied, not referenced.
	println(&sArr[2] == &newArr[1])
	// share same underlying array, and same index
	println(&sArr[1] == &newArr[1])
}

// Output:
// 1 2 3
// true
// 2 3 3
// true
// 2 3
// false
// true
