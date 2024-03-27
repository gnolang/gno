package main

type S struct {
	i int
}

func main() {
	sArr := make([]S, 0, 4)
	sArr = append(sArr, S{1}, S{2}, S{3})

	newArr := append(sArr[:0], sArr[0:]...)

	// share same underlying array
	println(&sArr[0] == &newArr[0])

	println(&sArr[1] == &newArr[1])

	println(&sArr[2] == &newArr[2])
}

// Output:
// true
// true
// true
