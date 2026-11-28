// Legitimate deep interface nesting: three distinct embedded layers resolve to a
// concrete method. Distinct operands never repeat, so detection does not fire
// (guards against being over-strict on depth). Converges to the concrete value.
package main

type IG interface{ Get() int }

type Concrete struct{ v int }

func (c Concrete) Get() int { return c.v }

type Mid struct{ IG }

type Outer struct{ IG }

func main() {
	var o IG = Outer{IG: Mid{IG: Concrete{v: 9}}}
	g := o.Get
	println(g())
}

// Output:
// 9
