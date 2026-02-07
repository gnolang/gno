package main


var s1 []*int

func inc2(j *int) {
	*j = *j + 2 // Just as an example, increment j by 2.
}

func forLoopRef() {
	defer func() {
		for i, e := range s1 {
			println("i, e: ", i, *e)
		}
	}()

	for i := 0; i < 3; inc2(&i) {
		s1 = append(s1, &i)
		for j := 0; j < 2; inc2(&j){
			s1 = append(s1, &j)
		}
	}
}

func main() {
	forLoopRef()
}

// go 1.22 loop var is not supported for now.

// Preprocessed:
// file{ package main; var s1<!~VPBlock(2,0)> []*((const-type int)); func inc2(j *((const-type int))) { *(j<VPBlock(1,0)>) = *(j<VPBlock(1,0)>) + (const (2 int)) }; func forLoopRef() { defer func func(){ for i<!VPBlock(1,0)>, e<!VPBlock(1,1)> := range (const (ref(main) package{})).s1 { (const (println func(...interface {})))((const ("i, e: " string)), i<VPBlock(1,0)>, *(e<VPBlock(1,1)>)) } }(); for .loopvar_i<!~VPBlock(1,0)> := (const (0 int)); .loopvar_i<~VPBlock(1,0)> < (const (3 int)); inc2<VPBlock(4,1)>(&(.loopvar_i<~VPBlock(1,0)>)) { .redefine_i<!~VPBlock(1,0)> := .loopvar_i<~VPBlock(2,0)>; s1<~VPBlock(5,0)> = (const (append func([]*int, ...*int) []*int))(s1<~VPBlock(5,0)>, &(.redefine_i<~VPBlock(1,0)>)); for .loopvar_j<!~VPBlock(1,0)> := (const (0 int)); .loopvar_j<~VPBlock(1,0)> < (const (2 int)); inc2<VPBlock(6,1)>(&(.loopvar_j<~VPBlock(1,0)>)) { .redefine_j<!~VPBlock(1,0)> := .loopvar_j<~VPBlock(2,0)>; s1<~VPBlock(7,0)> = (const (append func([]*int, ...*int) []*int))(s1<~VPBlock(7,0)>, &(.redefine_j<~VPBlock(1,0)>)); .loopvar_j<*~VPBlock(2,0)> = .redefine_j<~VPBlock(1,0)> } ; .loopvar_i<*~VPBlock(2,0)> = .redefine_i<~VPBlock(1,0)> }  }; func main() { forLoopRef<VPBlock(3,2)>() } }

// Output:
// i, e:  0 0
// i, e:  1 0
// i, e:  2 2
// i, e:  3 0
