package main

type Int int

var s1 []*Int

func inc2(j *Int) {
	*j = *j + 2 
	s1 = append(s1, j)
}

func forLoopRef() {
	for i := Int(0); i < 10; inc2(&i) {
		println("ok")

		func(){
			println(i)
		}()

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
// file{ package main; type Int (const-type main.Int); var s1<!~VPBlock(2,1)> []*(typeval{main.Int}); func inc2(j *(typeval{main.Int})) { *(j<VPBlock(1,0)>) = *(j<VPBlock(1,0)>) + (const (2 main.Int)); s1<~VPBlock(3,1)> = (const (append func([]*main.Int, ...*main.Int) []*main.Int))(s1<~VPBlock(3,1)>, j<VPBlock(1,0)>) }; func forLoopRef() { for .loopvar_i<!@VPBlock(1,0)> := (const (0 main.Int)); .loopvar_i<*VPBlock(1,0)> < (const (10 main.Int)); inc2<VPBlock(4,2)>(&(.loopvar_i<~VPBlock(1,0)>)) { .redefine_i<!~VPBlock(1,0)> := .loopvar_i<*VPBlock(2,0)>; (const (println func(...interface {})))((const ("ok" string))); func func(){ (const (println func(...interface {})))(.redefine_i<~VPBlock(1,0)>) }<.redefine_i<()~VPBlock(1,0)>>(); .loopvar_i<*~VPBlock(2,0)> = .redefine_i<~VPBlock(1,0)> } ; for i<!VPBlock(1,0)>, e<!VPBlock(1,1)> := range s1<~VPBlock(4,1)> { (const (println func(...interface {})))((const ("i, e: " string)), i<VPBlock(1,0)>, *(e<VPBlock(1,1)>)) } }; func main() { forLoopRef<VPBlock(3,3)>() } }

// Output:
// ok
// (0 main.Int)
// ok
// (2 main.Int)
// ok
// (4 main.Int)
// ok
// (6 main.Int)
// ok
// (8 main.Int)
// i, e:  0 (2 main.Int)
// i, e:  1 (4 main.Int)
// i, e:  2 (6 main.Int)
// i, e:  3 (8 main.Int)
// i, e:  4 (10 main.Int)
