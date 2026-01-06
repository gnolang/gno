package main

func main() {
	var f func()
OUTER:
	for i := 0; i < 3; {
		// .redefine_i = .loopvar_i
		println("enter, i: ", i)

		for j := 0; j < 2; j++ {
			f = func(){
				println("j: ", j)
			}
			if j == 1 {
				println("---continue")
				i++ // .redefine_i, => 3
				// .loopvar_i = .redefine_i
				continue OUTER
			}
		}

		i++

		f = func() {
			start := i
			println("capture, start: ", start)
		}
		f()
	}
}

// Preprocessed:
// file{ package main; func main() { var f<!VPBlock(1,0)> func(); for .loopvar_i<!VPBlock(1,0)> := (const (0 int)); .loopvar_i<VPBlock(1,0)> < (const (3 int));  { .redefine_i<!~VPBlock(1,0)> := .loopvar_i<VPBlock(2,0)>; (const (println func(...interface {})))((const ("enter, i: " string)), .redefine_i<~VPBlock(1,0)>); for .loopvar_j<!VPBlock(1,0)> := (const (0 int)); .loopvar_j<VPBlock(1,0)> < (const (2 int)); .loopvar_j<VPBlock(1,0)>++ { .redefine_j<!~VPBlock(1,0)> := .loopvar_j<VPBlock(2,0)>; f<VPBlock(5,0)> = func func(){ (const (println func(...interface {})))((const ("j: " string)), .redefine_j<~VPBlock(1,0)>) }<.redefine_j<()~VPBlock(1,0)>>; if .redefine_j<~VPBlock(2,0)> == (const (1 int)) { (const (println func(...interface {})))((const ("---continue" string))); .redefine_i<~VPBlock(4,0)>++; .loopvar_i<VPBlock(5,0)> = .redefine_i<~VPBlock(4,0)>; continue OUTER<0,0,0> }; .loopvar_j<VPBlock(2,0)> = .redefine_j<~VPBlock(1,0)> } ; .redefine_i<~VPBlock(1,0)>++; f<VPBlock(3,0)> = func func(){ start<!VPBlock(1,0)> := .redefine_i<~VPBlock(1,1)>; (const (println func(...interface {})))((const ("capture, start: " string)), start<VPBlock(1,0)>) }<.redefine_i<()~VPBlock(1,0)>>; f<VPBlock(3,0)>(); .loopvar_i<VPBlock(2,0)> = .redefine_i<~VPBlock(1,0)> }  } }

// Output:
// enter, i:  0
// ---continue
// enter, i:  1
// ---continue
// enter, i:  2
// ---continue
