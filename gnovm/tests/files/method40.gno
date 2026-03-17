// Test that direct method calls (OpMethodPrecall path) and stored method
// values (OpSelector+OpPrecall path via BoundMethodValue) produce identical
// results for value methods, pointer methods, interface dispatch, and
// embedded methods.

package main

type Counter struct {
	N int
}

func (c Counter) Value() int { return c.N }
func (c *Counter) Increment() { c.N++ }

type Stringer interface {
	String() string
}

type Named struct {
	Name string
}

func (n Named) String() string { return n.Name }

type Wrapper struct {
	Named
}

func main() {
	// Value method: direct vs stored
	c := Counter{N: 42}
	directVal := c.Value()
	storedValFn := c.Value
	storedVal := storedValFn()
	if directVal != storedVal {
		panic("value method mismatch")
	}
	println("value method:", directVal, storedVal)

	// Pointer method: direct vs stored
	c2 := &Counter{N: 10}
	c2.Increment()
	directPtrN := c2.N
	c3 := &Counter{N: 10}
	storedPtrFn := c3.Increment
	storedPtrFn()
	storedPtrN := c3.N
	if directPtrN != storedPtrN {
		panic("pointer method mismatch")
	}
	println("pointer method:", directPtrN, storedPtrN)

	// Interface dispatch: direct vs stored
	var s Stringer = Named{Name: "hello"}
	directStr := s.String()
	storedIfaceFn := s.String
	storedStr := storedIfaceFn()
	if directStr != storedStr {
		panic("interface method mismatch")
	}
	println("interface method:", directStr, storedStr)

	// Embedded method via interface: direct vs stored
	var s2 Stringer = Wrapper{Named{Name: "wrapped"}}
	directEmbed := s2.String()
	storedEmbedFn := s2.String
	storedEmbed := storedEmbedFn()
	if directEmbed != storedEmbed {
		panic("embedded interface method mismatch")
	}
	println("embedded method:", directEmbed, storedEmbed)
}

// Output:
// value method: 42 42
// pointer method: 11 11
// interface method: hello hello
// embedded method: wrapped wrapped
