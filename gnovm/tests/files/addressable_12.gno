package main

type E struct{ F int }

func (e *E) Inc() { e.F++ }

type S struct {
	*E
	V  E
	PA *[2]int
	SL []int
}

type P *E

type Arr [2]int

var gs = S{E: &E{}, PA: &[2]int{}, SL: make([]int, 2)}

func getS() S      { return gs }
func getP() P      { return P(gs.E) }
func getPE() *E    { return gs.E }
func getPS() *S    { return &gs }
func getPA() *Arr  { return (*Arr)(gs.PA) }
func getSl() []int { return gs.SL }

type W struct{ Arr [2]E }

func main() {
	// Field promoted through an embedded pointer of a call result.
	getS().F = 1
	getS().F++
	getS().F += 1
	getS().Inc()
	f := getS().Inc
	f()
	p := &getS().F
	*p += 1
	// Named pointer type.
	getP().F += 10
	// Call results that are pointers.
	getPE().F += 100
	getPE().Inc()
	getPS().V.F = 7
	getPS().V.Inc()
	q := &getPS().V
	q.F++
	// Index through a pointer-to-array and a slice of a call result.
	getS().PA[0] = 3
	getPA()[1] = 4
	getSl()[0] = 5
	getS().SL[1] = 6
	r := &getSl()[0]
	*r++
	// Composite literal address.
	w := &W{}
	w.Arr[0].Inc()
	println(gs.E.F, gs.V.F, gs.PA[0], gs.PA[1], gs.SL[0], gs.SL[1], w.Arr[0].F)
}

// Output:
// 117 9 3 4 6 6 1
