package main

// var a depends on both B() and C(), both of which read a.
// The cycle error should always name B first (B < C lexicographically).

var a = B() + C()

func B() int { return a }

func C() int { return a }

func main() {}

// Error:
// main/var_initorder_circ_multi.gno:6:5: circular dependency: a -> B -> a

// TypeCheckError:
// main/var_initorder_circ_multi.gno:6:5: initialization cycle for a; main/var_initorder_circ_multi.gno:6:5: 	a refers to B; main/var_initorder_circ_multi.gno:8:6: 	B refers to a
