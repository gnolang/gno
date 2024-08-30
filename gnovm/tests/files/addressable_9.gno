package main

type S struct {
	i int
	t *T
}

type T struct {
	i int
}

func main() {
	m := map[int]*S{}
	s := &S{i: 4}
	m[5] = s
	println(&m[5].i)

	mm := map[int]S{}
	ss := S{t: new(T)}
	mm[8] = ss
	println(&mm[8].t.i)

	mmm := map[int]map[int]*S{}
	mmm[3] = map[int]*S{}
	mmm[3][3] = &S{i: 7}
	println(&mmm[3][3].i)
}

// Output:
// &(4 int)
// &(0 int)
// &(7 int)
