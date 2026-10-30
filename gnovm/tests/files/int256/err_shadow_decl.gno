package main

type int256 struct{ v int }

func main() {
	println(int256{v: 1}.v)
}

// Error:
// main/err_shadow_decl.gno:3:6-28: builtin identifiers cannot be shadowed: int256

// TypeCheckError:
// main/.gnobuiltins.gno:9:6: int256 redeclared in this block; main/err_shadow_decl.gno:3:6: 	other declaration of int256
