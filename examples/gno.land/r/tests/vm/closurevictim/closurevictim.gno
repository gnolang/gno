// Package closurevictim runs a caller's function value over its own state.
package closurevictim

var data = []int{0}

// Run calls f, a value from the caller, on this realm's data.
func Run(cur realm, f func([]int)) { f(data) }

func Get() int { return data[0] }
