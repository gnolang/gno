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

// Preprocessed:
// file{ package main; func main() { c<!VPBlock(1,0)> := (const (0 int)); refs<!VPBlock(1,1)> := []*((const-type int)){}; var i<!~VPBlock(1,2)> (const-type int) = (const (1 int)); refs<VPBlock(1,1)> = (const (append func(x []*int,args ...*int)(res []*int)))(refs<VPBlock(1,1)>, &(i<~VPBlock(1,2)>)); i<~VPBlock(1,2)> += (const (1 int)); c<VPBlock(1,0)> += (const (1 int)); if c<VPBlock(2,0)> < (const (10 int)) { goto loop<1,2> }; for _<VPBlock(0,0)>, ref<VPBlock(1,0)> := range refs<VPBlock(2,1)> { (const (println func(xs ...interface{})()))(*(ref<VPBlock(1,0)>)) } } }

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
