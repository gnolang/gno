package main

func main() {
	var f func()
	for i := 0; i < 3;{
		println("enter, i: ", i)

		if i == 1 {
			println("---continue")
			i++ 
			continue
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
// file{ package main; func main() { var f<!VPBlock(1,0)> func(); for i.loopvar<!~VPBlock(1,0)> := (const (0 int)); i.loopvar<~VPBlock(1,0)> < (const (3 int));  { (const (println func(...interface {})))((const ("enter, i: " string)), i.loopvar<~VPBlock(1,0)>); if i.loopvar<~VPBlock(2,0)> == (const (1 int)) { (const (println func(...interface {})))((const ("---continue" string))); i.loopvar<~VPBlock(2,0)>++; continue }; i.loopvar<~VPBlock(1,0)>++; f<VPBlock(2,0)> = func func(){ start<!VPBlock(1,0)> := i.loopvar<~VPBlock(1,1)>; (const (println func(...interface {})))((const ("capture, start: " string)), start<VPBlock(1,0)>) }<i.loopvar<()~VPBlock(1,0)>>; f<VPBlock(2,0)>() } } }

// Output:
// enter, i:  0
// capture, start:  1
// enter, i:  1
// ---continue
// enter, i:  2
// capture, start:  3
