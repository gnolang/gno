package main

// Two unassigned interface{} variables compare equal. Note this goes
// through isEql's IsUndefined() early-return path (T=nil), not the
// `case InterfaceKind:` switch arm.
func main() {
	var x, y interface{}
	println(x == y)
}

// Output:
// true
