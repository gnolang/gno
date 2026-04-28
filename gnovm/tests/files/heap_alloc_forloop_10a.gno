package main

func main() {
	var f func()
OUTER:
	for i := 0; i < 3; {
		println("enter, i: ", i)

		for j := 0; j < 2; j++ {
			f = func(){
				println("j: ", j)
			}
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
// file{ package main; func main() { var f<!VPBlock(1,0)> func(); for i.loopvar<!~VPBlock(1,0)> := (const (0 int)); i.loopvar<~VPBlock(1,0)> < (const (3 int));  { (const (println func(...interface {})))((const ("enter, i: " string)), i.loopvar<~VPBlock(1,0)>); for j.loopvar<!~VPBlock(1,0)> := (const (0 int)); j.loopvar<~VPBlock(1,0)> < (const (2 int)); j.loopvar<~VPBlock(1,0)>++ { f<VPBlock(3,0)> = func func(){ (const (println func(...interface {})))((const ("j: " string)), j.loopvar<~VPBlock(1,0)>) }<j.loopvar<()~VPBlock(1,0)>>; if j.loopvar<~VPBlock(2,0)> == (const (1 int)) { (const (println func(...interface {})))((const ("---continue" string))); i.loopvar<~VPBlock(3,0)>++; continue OUTER<0,0,0> } }; i.loopvar<~VPBlock(1,0)>++; f<VPBlock(2,0)> = func func(){ start<!VPBlock(1,0)> := i.loopvar<~VPBlock(1,1)>; (const (println func(...interface {})))((const ("capture, start: " string)), start<VPBlock(1,0)>) }<i.loopvar<()~VPBlock(1,0)>>; f<VPBlock(2,0)>() } } }

// Output:
// enter, i:  0
// ---continue
// enter, i:  1
// ---continue
// enter, i:  2
// ---continue
