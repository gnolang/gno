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
	thing := func() {
		i := 2 // new i
		j := 3 // new j
		println(i)
		println(j)
	}
	if c < 10 {
		goto loop
	}
	for _, ref := range refs {
		println(*ref)
	}
	if false {
		println(j) // dummy usage
	}
	if false {
		thing() // dummy usage
	}
}

// This does make 'i' NameExprTypeHeapDefine.
// You can tell by the preprocess printout of i<!~...> and i<~...>.

// Preprocessed:
// file{ package main; func main() { c<!VPBlock(1,0)> := (const (0 int)); refs<!VPBlock(1,1)> := (const-type []*int){}; var i<!~VPBlock(1,2)>, j<!VPBlock(1,3)> (const-type int) = c<VPBlock(1,0)>, (const (2 int)); refs<VPBlock(1,1)> = (const (append func([]*int, ...*int) []*int))(refs<VPBlock(1,1)>, &(i<~VPBlock(1,2)>)); i<~VPBlock(1,2)> += (const (1 int)); j<VPBlock(1,3)> += (const (1 int)); c<VPBlock(1,0)> += (const (1 int)); thing<!VPBlock(1,4)> := func func(){ i<!VPBlock(1,0)> := (const (2 int)); j<!VPBlock(1,1)> := (const (3 int)); (const (println func(...interface {})))(i<VPBlock(1,0)>); (const (println func(...interface {})))(j<VPBlock(1,1)>) }; if c<VPBlock(2,0)> < (const (10 int)) { goto loop<1,0,2> }; for _<VPBlock(0,0)>, ref<!VPBlock(1,0)> := range refs<VPBlock(2,1)> { (const (println func(...interface {})))(*(ref<VPBlock(1,0)>)) }; if (const (false bool)) { (const (println func(...interface {})))(j<VPBlock(2,3)>) }; if (const (false bool)) { thing<VPBlock(2,4)>() } } }

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
