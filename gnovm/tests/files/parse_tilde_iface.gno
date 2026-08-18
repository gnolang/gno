package main

type T interface {
	~int
}

func main() {}

// On the type-check path the pre-type-check cost guard rejects `~` first, since
// cost() cannot count a type-set term (see checkNoUncountableGenerics). Go2Gno's
// rejection below still covers the paths that never run go/types. Both fire, so
// both are pinned.

// Error:
// parse_tilde_iface.gno:4:2: tilde operator is not permitted

// TypeCheckError:
// main/parse_tilde_iface.gno:4:2: interface approximation (~) terms are not supported (Gno targets go1.17)
