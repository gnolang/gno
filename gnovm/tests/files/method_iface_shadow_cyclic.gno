// A concrete method that shadows an embedded interface wins dispatch, so even a
// cyclic embed (s.IG = s) is never consulted and resolution converges to the
// concrete method. Guards against the detection firing on a non-dispatched
// cycle (must NOT panic).
package main

type IG interface{ Get() int }

type Box struct {
	IG
	v int
}

func (b *Box) Get() int { return b.v }

func main() {
	s := &Box{v: 7}
	s.IG = s // cyclic, but *Box.Get shadows it
	var o IG = s
	g := o.Get
	println(g())
}

// Output:
// 7
