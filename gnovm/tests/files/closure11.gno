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

// Preprocessed:
// file{ package main; type fooer (const-type main.fooer); var f<!~VPBlock(2,1)> typeval{main.fooer} = (const (1 int)); func trymake() .res.0 func() { switch v:=f<~VPBlock(4,1)> { case (const-type int): return func func(){ (const (println func(...interface {})))(v<~VPBlock(1,0)>) }<v<()~VPBlock(1,0)>>; default:  }; return func func(){  } }; func main() { trymake<VPBlock(3,2)>()() } }

// Output:
// 1
