package main

// Tests init order with multiple vars all depending on the same function,
// which itself depends on a package-level var (transitive dep through function).
// Vars without unsatisfied deps are initialized in declaration order.
// Detailed order:
//   - events: no deps, initialized first (zero value, no emit call).
//   - B, A, C, G, D, E: each only calls emit(events), all ready after events.
//     Initialized in declaration order: B, A, C, G, D, E.
//   - Z: depends directly on A, B, C, D, E — all now satisfied. Z is placed
//     earlier in declaration order than F, so Z runs next (emits "L").
//   - F: last remaining, initialized after Z.

var events []string

func emit(s string) string { events = append(events, s); return s }

var (
	Z = A + "-" + B + "-" + C + "-" + emit("L") + D + "-" + E
	B = emit("B")
	A = emit("A")
	C = emit("C")
	G = emit("G")
	D = emit("D")
	E = emit("E")
	F = emit("F")
)

func main() {
	for _, e := range events {
		println(e)
	}
}

// Output:
// B
// A
// C
// G
// D
// E
// L
// F
