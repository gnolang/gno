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

// Preprocessed:
// file{ package main; func main() { c<!VPBlock(1,0)> := (const (0 int)); i<!VPBlock(1,1)> := (const (1 int)); (const (println func(...interface {})))(i<VPBlock(1,1)>); c<VPBlock(1,0)> += (const (1 int)); if c<VPBlock(2,0)> < (const (10 int)) { goto loop<1,0,1> } } }



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
