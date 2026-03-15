package main

// Tests that init-order analysis is not performed through interface dispatch,
// even when both variables call the same method via the same interface.
// Neither A nor B can have their dependency on T.Get (which reads B) inferred,
// because the static receiver is the Getter interface, not a *DeclaredType.
// Both are treated as ready immediately; they are initialized in declaration
// order (A first, then B), so A = B (zero) = 0.
//
// Note: the real Go compiler sees through the interface conversion and produces
// a different result. That behaviour is not required by the Go spec; Gno's
// result of 0 is spec-compliant.

type Getter interface {
	Get() int
}

type T struct{}

func (T) Get() int { return B }

var (
	A = Getter(T{}).Get()
	B = Getter(T{}).Get()
)

func main() { println(A) }

// Output:
// 0
