package main

type printer struct {
}

func newPrinter() *printer {
	return &printer{}
}

func (p *printer) doPrintf() {
	for i := 0; i < 3; i++{
		digits := func() string {
			start := i
			println(start)
			return "ok"
		}
		digits()
		// i++
		// println(i)
	}
}

func main() {
	p := newPrinter()
	p.doPrintf()
}

// Output:
// 0
// 1
// 2
