package main

var s1 []*int

func inc2(j *int) {
	*j = *j + 1
	s1 = append(s1, j)
}

func forLoopRef() {
	for i := 0; i < 5; inc2(&i) {
		if i == 2 {
		// Copy out.
			continue
		}

		println("ok")
		// Copy out.
	}

	println("s1: ", s1)
	for i, e := range s1 {
		println("i, e: ", i, *e)
	}
}

func main() {
	forLoopRef()
}

// Preprocessed:
// file{ package main; var s1<!~VPBlock(2,0)> []*((const-type int)); func inc2(j *((const-type int))) { *(j<VPBlock(1,0)>) = *(j<VPBlock(1,0)>) + (const (1 int)); s1<~VPBlock(3,0)> = (const (append func([]*int, ...*int) []*int))(s1<~VPBlock(3,0)>, j<VPBlock(1,0)>) }; func forLoopRef() { for .loopvar_i<!~VPBlock(1,0)> := (const (0 int)); .loopvar_i<~VPBlock(1,0)> < (const (5 int)); inc2<VPBlock(4,1)>(&(.loopvar_i<~VPBlock(1,0)>)) if .loopvar_i<~VPBlock(2,0)> == (const (2 int)) { .loopvar_i<*~VPBlock(2,0)> = .loopvar_i<~VPBlock(2,0)>; continue }; (const (println func(...interface {})))((const ("ok" string))); .loopvar_i<*~VPBlock(1,0)> = .loopvar_i<~VPBlock(1,0)> ; (const (println func(...interface {})))((const ("s1: " string)), s1<~VPBlock(3,0)>); for i<!VPBlock(1,0)>, e<!VPBlock(1,1)> := range s1<~VPBlock(4,0)> { (const (println func(...interface {})))((const ("i, e: " string)), i<VPBlock(1,0)>, *(e<VPBlock(1,1)>)) } }; func main() { forLoopRef<VPBlock(3,2)>() } }

// Output:
// ok
// ok
// ok
// ok
// s1:  slice[(&(1 int) *int),(&(2 int) *int),(&(3 int) *int),(&(4 int) *int),(&(5 int) *int)]
// i, e:  0 1
// i, e:  1 2
// i, e:  2 3
// i, e:  3 4
// i, e:  4 5
