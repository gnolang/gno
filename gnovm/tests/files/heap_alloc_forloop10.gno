package main


var f1s []func()
var f2s []func()

func main() {
	for i := 0; i < 3; i++ {
		f1 := func() {
			println("i: ", i)
		}

		f1s = append(f1s, f1)

		var i uint64 = uint64(i)

		f2 := func() {
			println("i: ", i)
		}

		f2s = append(f2s, f2)

	}
	for _, f1 := range f1s {
		f1()
	}

	for _, f2 := range f2s {
		f2()
	}
}

// Preprocessed:
// file{ package main; var f1s<!~VPBlock(2,0)> []func(); var f2s<!~VPBlock(2,1)> []func(); func main() { for .loopvar_i<!~VPBlock(1,0)> := (const (0 int)); .loopvar_i<~VPBlock(1,0)> < (const (3 int)); .loopvar_i<~VPBlock(1,0)>++ f1<!VPBlock(1,1)> := func func(){ (const (println func(...interface {})))((const ("i: " string)), .loopvar_i<~VPBlock(1,0)>) }<.loopvar_i<()~VPBlock(1,0)>>; f1s<~VPBlock(4,0)> = (const (append func([]func(), ...func()) []func()))(f1s<~VPBlock(4,0)>, f1<VPBlock(1,1)>); var i<!~VPBlock(1,2)> (const-type uint64) = (const-type uint64)(.loopvar_i<~VPBlock(1,0)>); f2<!VPBlock(1,3)> := func func(){ (const (println func(...interface {})))((const ("i: " string)), i<~VPBlock(1,0)>) }<i<()~VPBlock(1,2)>>; f2s<~VPBlock(4,1)> = (const (append func([]func(), ...func()) []func()))(f2s<~VPBlock(4,1)>, f2<VPBlock(1,3)>); .loopvar_i<*~VPBlock(1,0)> = .loopvar_i<~VPBlock(1,0)> ; for _<VPBlock(0,0)>, f1<!VPBlock(1,0)> := range f1s<~VPBlock(4,0)> { f1<VPBlock(1,0)>() }; for _<VPBlock(0,0)>, f2<!VPBlock(1,0)> := range f2s<~VPBlock(4,1)> { f2<VPBlock(1,0)>() } } }

// Output:
// i:  0
// i:  1
// i:  2
// i:  0
// i:  1
// i:  2
