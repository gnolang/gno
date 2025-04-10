// Package emit demonstrates how to use the std.Emit() function
// to emit Gno events that can be used to track data changes off-chain.
// std.Emit is variadic; apart from the event name, it can take in any number of key-value pairs to emit.
package emit

import "chain"

func Emit(value string) {
	chain.Emit("EventName", "key", value)
}
