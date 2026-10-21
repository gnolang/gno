package main

// A pointer method may be called on an addressable receiver; the VM
// synthesizes the &receiver. Every form below is addressable, so all of
// them must keep working.

type T struct{ n int }

func (t *T) Inc() { t.n++ }

type Embed struct{ T }

type Field struct{ in T }

type I struct{ z int }

func (i *I) Bump() { i.z++ }

type EmbedPtr struct{ *I }

func mkPtr() EmbedPtr { return EmbedPtr{&I{}} }

func main() {
	v := T{}
	v.Inc()

	p := &T{}
	p.Inc()

	s := []T{{}}
	s[0].Inc()

	a := [1]T{}
	a[0].Inc()

	e := Embed{}
	e.Inc()

	f := Field{}
	f.in.Inc()

	// reached through an embedded pointer, so the outer value need not
	// be addressable itself.
	mkPtr().Bump()

	println(v.n, p.n, s[0].n, a[0].n, e.n, f.in.n)
}

// Output:
// 1 1 1 1 1 1
