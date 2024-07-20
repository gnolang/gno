package main

func main() {
	c := 0
	refs := []*int{}
loop:
	var i, j int = c, 2
	refs = append(refs, &i)
	i += 1
	j += 1
	c += 1
	if c < 10 {
		goto loop
	}
	for _, ref := range refs {
		println(*ref)
	}
	if false {
		println(j) // dummy usage
	}
}

// This does make 'i' NameExprTypeHeapDefine.
// You can tell by the preprocess printout of i<!~...> and i<~...>.

// Output:
// 1
// 2
// 3
// 4
// 5
// 6
// 7
// 8
// 9
// 10
