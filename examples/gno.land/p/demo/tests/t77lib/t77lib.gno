// Package t77lib passes a callback on down chains of functions.
package t77lib

// Each passes f on to each2, which only calls it.
func Each(xs []int, f func(int)) { each2(xs, f) }

func each2(xs []int, f func(int)) {
	for _, x := range xs {
		f(x)
	}
}

// One calls f itself: the one-hop twin.
func One(xs []int, f func(int)) {
	for _, x := range xs {
		f(x)
	}
}

// Three passes f on twice.
func Three(xs []int, f func(int)) { three2(xs, f) }

func three2(xs []int, f func(int)) { each2(xs, f) }

type Node struct {
	Left, Right *Node
	V           int
}

// Walk visits n in order, passing f to itself.
func Walk(n *Node, f func(int)) {
	if n == nil {
		return
	}
	Walk(n.Left, f)
	f(n.V)
	Walk(n.Right, f)
}

// Ping and pong pass f to each other.
func Ping(n int, f func(int)) {
	if n <= 0 {
		return
	}
	f(n)
	pong(n-1, f)
}

func pong(n int, f func(int)) { Ping(n, f) }

// Holder keeps a callback.
type Holder struct{ F func(int) }

// Stash passes f on to a function that stores it.
func Stash(h *Holder, f func(int)) { stash2(h, f) }

func stash2(h *Holder, f func(int)) { h.F = f }

type Inner struct{}

// Do passes f on to Each.
func (Inner) Do(xs []int, f func(int)) { Each(xs, f) }
