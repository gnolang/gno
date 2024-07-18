package main

func main() {
	c := 0
	closures := []func(){}
loop:
	i := c
	closures = append(closures, func() {
		println(i)
	})
	c += 1
	if c < 10 {
		goto loop
	}

	for _, cl := range closures {
		cl()
	}
}

// This does make 'i' NameExprTypeHeapDefine.
// You can tell by the preprocess printout of i<!~...> and i<()~...>.

// Output:
// 0
// 1
// 2
// 3
// 4
// 5
// 6
// 7
// 8
// 9
