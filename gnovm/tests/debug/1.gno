package main

type S struct {
	i int
}

func main() {
	sArr := make([]S, 0, 4)
	sArr = append(sArr, S{1}, S{2}, S{3})

	newArr := append(sArr[:0], sArr[1:]...)

	// The struct should have been copied, not referenced.
	println(&sArr[2] == &newArr[1])
}

// Output:
// false
