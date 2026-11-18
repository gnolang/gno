package main

// A for-init that shadows an outer variable and reads it on the RHS:
// the RHS x is the outer x (evaluated in the enclosing scope), so it must
// not be renamed to x.loopvar. The closure after the loop captures the
// outer x too. Previously the loopvar pass renamed the RHS x and preprocess
// failed with "name x.loopvar not declared".
func main() {
	x := 7
	for x, i := x, 0; i < 1; i++ {
		x = 0
		_ = x
	}
	f := func() { println(x) }
	f()
}

// Preprocessed:
// file{ package main; func main() { x<!~VPBlock(1,0)> := (const (7 int)); for x.loopvar<!VPBlock(1,0)>, i.loopvar<!VPBlock(1,1)> := x<~VPBlock(2,0)>, (const (0 int)); i.loopvar<VPBlock(1,1)> < (const (1 int)); i.loopvar<VPBlock(1,1)>++ { x.loopvar<VPBlock(1,0)> = (const (0 int)); _<VPBlock(0,0)> = x.loopvar<VPBlock(1,0)> }; f<!VPBlock(1,1)> := func func(){ (const (println func(...interface {})))(x<~VPBlock(1,0)>) }<x<()~VPBlock(1,0)>>; f<VPBlock(1,1)>() } }

// Output:
// 7
