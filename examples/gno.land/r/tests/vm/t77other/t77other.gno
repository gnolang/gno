// Package t77other is another realm whose plain function passes f on.
package t77other

func Each(xs []int, f func(int)) { each2(xs, f) }

func each2(xs []int, f func(int)) {
	for _, x := range xs {
		f(x)
	}
}

// E is another realm's type whose method passes f on.
type E struct{}

func (*E) Each(xs []int, f func(int)) { each2(xs, f) }

// NewE returns an E made by this realm.
func NewE() any { return &E{} }
