package main

func main() {
	var f func()
OUTER:
	for i := 0; i < 3; {
		println("enter, i: ", i)

		for j := 0; j < 2; j++ {
			if j == 1 {
				println("---continue")
				i++
				// Copy out.
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
// file{ package main; func main() { var f<!VPBlock(1,0)> func(); for .loopvar_i<!~VPBlock(1,0)> := (const (0 int)); .loopvar_i<~VPBlock(1,0)> < (const (3 int));  (const (println func(...interface {})))((const ("enter, i: " string)), .loopvar_i<~VPBlock(1,0)>); for .loopvar_j<!VPBlock(1,0)> := (const (0 int)); .loopvar_j<VPBlock(1,0)> < (const (2 int)); .loopvar_j<VPBlock(1,0)>++ if .loopvar_j<VPBlock(2,0)> == (const (1 int)) { (const (println func(...interface {})))((const ("---continue" string))); .loopvar_i<~VPBlock(3,0)>++; .loopvar_i<@!~VPBlock(3,0)> = .loopvar_i<~VPBlock(3,0)>; continue OUTER<0,0,0> } ; .loopvar_i<~VPBlock(1,0)>++; f<VPBlock(2,0)> = func func(){ start<!VPBlock(1,0)> := .loopvar_i<~VPBlock(1,1)>; (const (println func(...interface {})))((const ("capture, start: " string)), start<VPBlock(1,0)>) }<.loopvar_i<()~VPBlock(1,0)>>; f<VPBlock(2,0)>(); .loopvar_i<@!~VPBlock(1,0)> = .loopvar_i<~VPBlock(1,0)>  } }

// Output:
// enter, i:  0
// ---continue
// enter, i:  1
// ---continue
// enter, i:  2
// ---continue
