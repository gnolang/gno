package main

type fooer interface{}

var f fooer = 1

func trymake() func() {
	switch v := f.(type) {
	case int:
		return func() {
			println(v)
		}
	default:
	}
	return func() {}
}
func main() {
	trymake()()
}

// XXX Fix this by changing Varname to a *NameExpr
// and preprocessing similarly to an expression.

// Preprocessed:
// file{ package main; type fooer (const-type main.fooer); var f<!~VPBlock(2,1)> fooer<VPBlock(2,0)> = (const (1 int)); func trymake() .res.0 func() { switch v<!~VPBlock(1,0)> :=f<~VPBlock(4,1)> { case (const-type int): return func func(){ (const (println func(...interface {})))(v<~VPBlock(1,0)>) }<v<()~VPBlock(1,0)>>; default:  }; return func func(){  } }; func main() { trymake<VPBlock(3,2)>()() } }

// Error:
// should not happen, should be heapItemType: v<()~VPBlock(1,0)>
