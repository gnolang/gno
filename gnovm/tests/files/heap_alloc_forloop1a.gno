package main

type Int int

var s1 []*Int

func inc2(j *Int) {
	*j = *j + 2 // Just as an example, increment j by 2.
}

func forLoopRef() {
	defer func() {
		for i, e := range s1 {
			println("i, e: ", i, *e)
		}
	}()

	for i := Int(0); i < 10; inc2(&i) {
		// in rename, after inject, search depth to depth - 1
		s1 = append(s1, &i)
	}
}

func main() {
	forLoopRef()
}

// go 1.22 loop var is not supported for now.

// Preprocessed:
// file{ package main; type Int (const-type main.Int); var s1<!~VPBlock(2,1)> []*(typeval{main.Int}); func inc2(j *(typeval{main.Int})) { *(j<VPBlock(1,0)>) = *(j<VPBlock(1,0)>) + (const (2 main.Int)) }; func forLoopRef() { defer func func(){ for i<!VPBlock(1,0)>, e<!VPBlock(1,1)> := range (const (ref(main) package{})).s1 { (const (println func(...interface {})))((const ("i, e: " string)), i<VPBlock(1,0)>, *(e<VPBlock(1,1)>)) } }(); for .loopvar_i<!~VPBlock(1,0)> := (const (0 main.Int)); .loopvar_i<~VPBlock(1,0)> < (const (10 main.Int)); inc2<VPBlock(4,2)>(&(.loopvar_i<~VPBlock(1,0)>)) { .redefine_i<!~VPBlock(1,0)> := .loopvar_i<~VPBlock(2,0)>; s1<~VPBlock(5,1)> = (const (append func([]*main.Int, ...*main.Int) []*main.Int))(s1<~VPBlock(5,1)>, &(.redefine_i<~VPBlock(1,0)>)); .loopvar_i<*~VPBlock(2,0)> = .redefine_i<~VPBlock(1,0)> }  }; func main() { forLoopRef<VPBlock(3,3)>() } }

// Output:
// i, e:  0 (0 main.Int)
// i, e:  1 (2 main.Int)
// i, e:  2 (4 main.Int)
// i, e:  3 (6 main.Int)
// i, e:  4 (8 main.Int)
