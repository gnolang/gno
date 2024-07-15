package main

func main() {
	c := 0
	refs := []*int{}
loop:
	var i int = 1
	refs = append(refs, &i)
	i += 1
	c += 1
	if c < 10 {
		goto loop
	}
	for _, ref := range refs {
		println(*ref)
	}
}

// This does make 'i' NameExprTypeHeapDefine.
// You can tell by the preprocess printout of i<!~...> and i<~...>.

// Output:
// 2
// 2
// 2
// 2
// 2
// 2
// 2
// 2
// 2
// 2
