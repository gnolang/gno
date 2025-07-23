package main

var a = 1

func trymake() func() {
	switch a {
	case 1:
		ok := true
		return func() {
			println(ok)
		}
	default:
	}
	return func() {}
}
func main() {
	trymake()()
}

// Preprocessed:
// file{ package main; var a<!~VPBlock(2,0)> = (const (1 int)); func trymake() .res.0 func() { switch a<~VPBlock(4,0)> { case (const (1 int)): ok<!~VPBlock(1,0)> := (const (true bool)); return func func(){ (const (println func(...interface {})))(ok<~VPBlock(1,0)>) }<ok<()~VPBlock(1,0)>>; default:  }; return func func(){  } }; func main() { trymake<VPBlock(3,1)>()() } }

// Output:
// true
