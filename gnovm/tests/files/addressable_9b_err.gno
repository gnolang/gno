package main

type S struct {
	i int
}

func main() {
	mmm := map[int]map[int]S{}
	mmm[3] = map[int]S{}
	mmm[3][3] = S{i: 7}
	println(&mmm[3][3].i)
}

// Output:
// &(7 int)

// TypeCheckError:
// main/files/addressable_9b_err.gno:11:11: invalid operation: cannot take address of mmm[3][3].i (value of type int)
