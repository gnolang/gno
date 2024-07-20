package main

func main() {
	c := 0
loop:
	i := 1
	println(i)
	c += 1
	if c < 10 {
		goto loop
	}
}

// This does not make 'i' NameExprTypeHeapDefine,
// because it is not actually used in a &ref or closure context.

// Output:
// 1
// 1
// 1
// 1
// 1
// 1
// 1
// 1
// 1
// 1
