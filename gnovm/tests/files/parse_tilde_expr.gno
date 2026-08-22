package main

var x = ~0

func main() {}

// Error:
// parse_tilde_expr.gno:3:9: tilde operator is not permitted

// TypeCheckError:
// main/parse_tilde_expr.gno:3:9: cannot use ~ outside of interface or type constraint (use ^ for bitwise complement)
