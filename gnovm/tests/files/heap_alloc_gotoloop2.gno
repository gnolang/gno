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

// Preprocessed:
// file{ package main; func main() { c<!VPBlock(1,0)> := (const (0 int)); closures<!VPBlock(1,1)> := (const-type []func()){}; i<!~VPBlock(1,2)> := c<VPBlock(1,0)>; closures<VPBlock(1,1)> = (const (append func([]func(), ...func()) []func()))(closures<VPBlock(1,1)>, func func(){ (const (println func(...interface {})))(i<~VPBlock(1,0)>) }<i<()~VPBlock(1,2)>>); c<VPBlock(1,0)> += (const (1 int)); if c<VPBlock(2,0)> < (const (10 int)) { goto loop<1,0,2> }; for _<VPBlock(0,0)>, cl<!VPBlock(1,0)> := range closures<VPBlock(2,1)> { cl<VPBlock(1,0)>() } } }

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
