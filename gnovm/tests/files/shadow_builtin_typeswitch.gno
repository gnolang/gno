package main

// The type-switch var is checked at the switch statement, so a switch with
// no clauses, which never reaches a clause's Reserve, is refused too.

func main() {
	switch len := any(1).(type) {
	}
	println("ok")
}

// Error:
// main/shadow_builtin_typeswitch.gno:7:2-8:3: builtin identifiers cannot be shadowed: len

// TypeCheckError:
// main/shadow_builtin_typeswitch.gno:7:9: len declared and not used
