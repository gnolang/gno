package main

type W[P any] struct {
	a, b [0]P
}

func main() {
	println("unreachable")
}

// The pre-type-check cost guard rejects the type-parameter declaration before
// go/types runs (TypeCheckError below): cost() scores an instantiation by its
// base type and drops the type arguments, so a fan-out could hide there. That is
// what stops a generic type-expansion fan-out from reaching validType. The
// secondary `name P not defined` from preprocess is the unrelated follow-on
// error described in adr/pr5826_typecheck_dos_guards.md.

// Error:
// main/type_param0.gno:3:6-5:2: name P not defined in fileset with files [type_param0.gno]

// TypeCheckError:
// main/type_param0.gno:3:6: generic type declarations are not supported
