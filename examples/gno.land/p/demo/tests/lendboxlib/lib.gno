package lendboxlib

type Box struct{ N int }

func Bump(b *Box) { b.N += 1000 }

func BumpInt(p *int) { *p += 1000 }

type Setter struct{ V int }

func (s Setter) Apply(b *Box) { b.N = s.V }

func Each(bs []*Box, f func(*Box)) {
	for _, b := range bs {
		f(b)
	}
}

func EachInt(p *int, f func(*int)) { f(p) }

func Run(f func()) { f() }

func MakeAdder(d int) func(*Box) { return func(b *Box) { b.N += d } }

func MakeIntAdder(d int) func(*int) { return func(p *int) { *p += d } }

// Guard wraps a callback (a typical middleware combinator).
func Guard(f func(*Box)) func(*Box) {
	return func(b *Box) {
		if b != nil {
			f(b)
		}
	}
}

func GuardInt(f func(*int)) func(*int) { return func(p *int) { f(p) } }
