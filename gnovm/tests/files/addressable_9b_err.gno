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

// Error:
// main/files/addressable_9b_err.gno:11:10: cannot take address of mmm<VPBlock(1,0)>[(const (3 int))][(const (3 int))].i
