package main

func main() {
	c := 0
	closures := []func(){}
	goto loop1
loop1: // not a loop
	i := 1
loop2:
	closures = append(closures, func() {
		println(i)
	})
	c += 1
	if c < 10 {
		goto loop2
	}
	for _, cl := range closures {
		cl()
	}
}

// This one doesn't because the goto stmt doesn't go back far enough.

// Preprocessed:
// file{ package main; func main() { c<!VPBlock(1,0)> := (const (0 int)); closures<!VPBlock(1,1)> := (const-type []func()){}; goto loop1<0,0,3>; i<!~VPBlock(1,2)> := (const (1 int)); closures<VPBlock(1,1)> = (const (append func([]func(), ...func()) []func()))(closures<VPBlock(1,1)>, func func(){ (const (println func(...interface {})))(i<~VPBlock(1,0)>) }<i<()~VPBlock(1,2)>>); c<VPBlock(1,0)> += (const (1 int)); if c<VPBlock(2,0)> < (const (10 int)) { goto loop2<1,0,4> }; for _<VPBlock(0,0)>, cl<!VPBlock(1,0)> := range closures<VPBlock(2,1)> { cl<VPBlock(1,0)>() } } }

// Output:
// 1
// 1
// 1
// 1
// 1
// 1
// 1
// 1
// 1
// 1
