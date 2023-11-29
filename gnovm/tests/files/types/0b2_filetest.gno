package main

type Error string

func (e Error) Error() string {
	return "error: " + string(e)
}

// left is untyped const, right is typed const
// untyped value can be converted to its correspondence value when:
// a) it's (untyped) const
// b) it's declared type of value, like type Error int8, 0 will be able to converted to Error
func main() {
	if 1 == Error(1) {
		println("what the firetruck?")
	} else {
		println("something else")
	}
}

// Error:
// main/files/types/0b2_filetest.gno:14: cannot convert untyped bigint type to StringKind
