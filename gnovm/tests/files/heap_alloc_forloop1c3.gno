package main

type Int int

var s1 []*Int


func forLoopRef() {
	for i := Int(0); i < 10; i++ {
		s1 = append(s1, &i)
		// .loopvar_i<loopvarHeapUse> = .redefine_i
	}

	for i, e := range s1 {
		println("i, e: ", i, *e)
	}
}

func main() {
	forLoopRef()
}

// Preprocessed:
// file{ package main; type Int (const-type main.Int); var s1<!~VPBlock(2,1)> []*(typeval{main.Int}); func forLoopRef() { for .loopvar_i<!~VPBlock(1,0)> := (const (0 main.Int)); .loopvar_i<~VPBlock(1,0)> < (const (10 main.Int)); .loopvar_i<~VPBlock(1,0)>++ s1<~VPBlock(4,1)> = (const (append func([]*main.Int, ...*main.Int) []*main.Int))(s1<~VPBlock(4,1)>, &(.loopvar_i<~VPBlock(1,0)>)); .loopvar_i<*~VPBlock(1,0)> = .loopvar_i<~VPBlock(1,0)> ; for i<!VPBlock(1,0)>, e<!VPBlock(1,1)> := range s1<~VPBlock(4,1)> { (const (println func(...interface {})))((const ("i, e: " string)), i<VPBlock(1,0)>, *(e<VPBlock(1,1)>)) } }; func main() { forLoopRef<VPBlock(3,2)>() } }

// Output:
// i, e:  0 (0 main.Int)
// i, e:  1 (1 main.Int)
// i, e:  2 (2 main.Int)
// i, e:  3 (3 main.Int)
// i, e:  4 (4 main.Int)
// i, e:  5 (5 main.Int)
// i, e:  6 (6 main.Int)
// i, e:  7 (7 main.Int)
// i, e:  8 (8 main.Int)
// i, e:  9 (9 main.Int)
