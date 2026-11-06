package main

// A range that assigns (not defines) an enclosing for-loop's vars: the key
// and value are references to k.loopvar / v.loopvar, not declarations, and
// must be renamed like any other reference. Previously the range key/value
// position was always treated as a declaration site, and preprocess failed
// with "name k not declared".
func main() {
	var fns []func() int
	for k, v := 0, 0; k < 1; k++ {
		for k, v = range []int{7, 8} {
			fns = append(fns, func() int { return k*10 + v })
		}
		println(k, v)
	}
	for _, f := range fns {
		println(f())
	}
}

// Preprocessed:
// file{ package main; func main() { var fns<!VPBlock(1,0)> []func() .res.0 (const-type int); for k.loopvar<!~VPBlock(1,0)>, v.loopvar<!~VPBlock(1,1)> := (const (0 int)), (const (0 int)); k.loopvar<~VPBlock(1,0)> < (const (1 int)); k.loopvar<~VPBlock(1,0)>++ { for k.loopvar<~VPBlock(2,0)>, v.loopvar<~VPBlock(2,1)> = range (const-type []int){(const (7 int)), (const (8 int))} { fns<VPBlock(3,0)> = (const (append func([]func() int, ...func() int) []func() int))(fns<VPBlock(3,0)>, func func() .res.0 (const-type int){ return k.loopvar<~VPBlock(1,1)> * (const (10 int)) + v.loopvar<~VPBlock(1,2)> }<k.loopvar<()~VPBlock(2,0)>, v.loopvar<()~VPBlock(2,1)>>) }; (const (println func(...interface {})))(k.loopvar<~VPBlock(1,0)>, v.loopvar<~VPBlock(1,1)>) }; for _<VPBlock(0,0)>, f<!VPBlock(1,0)> := range fns<VPBlock(2,0)> { (const (println func(...interface {})))(f<VPBlock(1,0)>()) } } }

// Output:
// 1 8
// 18
// 18
