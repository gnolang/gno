// Concrete-vs-interface receiver timing (matching Go): a concrete method value
// snapshots its receiver at bind (concrete 1); an interface method value derefs
// at the call, so a post-bind mutation is seen (snapshot 2), each call copies
// the receiver afresh (percall 11 11), and the value receiver cannot mutate the
// pointee (caller 10).

package main

type T struct{ x int }

func (t T) Get() int    { return t.x }
func (t T) IncGet() int { t.x++; return t.x }

type I interface {
	Get() int
	IncGet() int
}

func main() {
	// concrete: receiver copied at bind, later mutation invisible.
	c := &T{x: 1}
	cg := c.Get
	c.x = 2
	println("concrete", cg())

	// interface: receiver dereferenced at the call, mutation visible.
	p := &T{x: 1}
	var i I = p
	h := i.Get
	p.x = 2
	println("snapshot", h())

	// per-call copy + value semantics: each call starts from the current
	// pointee; the method's mutation does not leak back to it.
	q := &T{x: 10}
	var j I = q
	g := j.IncGet
	println("percall", g(), g())
	println("caller", q.x)
}

// Output:
// concrete 1
// snapshot 2
// percall 11 11
// caller 10
