package main

type T struct {
	a int
}

func (tv T) Mv(a int) int { return 0 }

func main() {
	const t = T.Mv
}

// Error:
// main/files/const39.gno:10:8: typeval{main.T}.Mv (variable of type func(main.T, int) int) is not constant

// TypeCheckError:
// main/files/const39.gno:10:12: T.Mv (value of type func(tv T, a int) int) is not constant
