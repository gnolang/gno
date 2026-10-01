package main

// The slice makes B's first appearance indirect, but the second field holds
// a B inline, so B has no finite size. The verdict must not depend on the
// direct reference being reached before the indirect one: by the time the
// second field is looked at, B is already defined.

type A = struct {
	s []B
	b B
}

type B A

func main() {}

// Error:
// main/recursive9i.gno:8:6-11:2: invalid recursive type: B refers to itself

// TypeCheckError:
// main/recursive9i.gno:13:6: invalid recursive type: B refers to itself
