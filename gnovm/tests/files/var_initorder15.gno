package main

// Tests that init-order analysis through interface dispatch is not performed.
// A = Getter(T{}).GetB() converts T{} to the Getter interface before calling
// GetB; at that point the static receiver type is *InterfaceType, not
// *DeclaredType, so the dependency on T.GetB (and transitively on B) cannot
// be inferred. A is therefore treated as having no deps and is initialized
// before B, so A = B (zero) = 0.
//
// Note: the real Go compiler performs a more sophisticated analysis that can
// see through interface conversions in simple cases like this and produces 42.
// That behaviour is not required by the Go spec, which only mandates stepwise
// initialization; Gno's result of 0 is spec-compliant.

type Getter interface {
	GetB() int
}

type T struct{}

func (T) GetB() int { return B }

var (
	A = Getter(T{}).GetB()
	B = 42
)

func main() { println(A) }

// Output:
// 0
