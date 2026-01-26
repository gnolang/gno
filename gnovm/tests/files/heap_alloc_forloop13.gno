package main

type printer struct {
}

func newPrinter() *printer {
	return &printer{}
}

func (p *printer) doPrintf() {
	for i := 0; i < 3;{
		digits := func() string {
			start := i
			println(start)
			return "ok"
		}
		digits()
		i++
	}
}

func main() {
	p := newPrinter()
	p.doPrintf()
}

// Preprocessed:
// file{ package main; type printer (const-type main.printer); func newPrinter() .res.0 *(typeval{main.printer}) { return &((const-type main.printer){}) }; func (p *(typeval{main.printer})) doPrintf() { for .loopvar_i<!~VPBlock(1,0)> := (const (0 int)); .loopvar_i<~VPBlock(1,0)> < (const (3 int));  digits<!VPBlock(1,1)> := func func() .res.0 (const-type string){ start<!VPBlock(1,1)> := .loopvar_i<~VPBlock(1,2)>; (const (println func(...interface {})))(start<VPBlock(1,1)>); return (const ("ok" string)) }<.loopvar_i<()~VPBlock(1,0)>>; digits<VPBlock(1,1)>(); .loopvar_i<~VPBlock(1,0)>++; .loopvar_i<*~VPBlock(1,0)> = .loopvar_i<~VPBlock(1,0)>  }; func main() { p<!VPBlock(1,0)> := newPrinter<VPBlock(3,1)>(); p<VPBlock(1,0)>.doPrintf() } }

// Output:
// 0
// 1
// 2
