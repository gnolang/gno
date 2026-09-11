package main

// An if branch is a nested scope, so it may shadow a name declared by the
// if's init statement, in either branch.
func main() {
	if v := 1; v == 1 {
		v := "then"
		println(v)
	}

	if v := 1; v == 2 {
		v := "then"
		println(v)
	} else {
		println("pre", v)
		v := "else"
		println("post", v)
	}
}

// Output:
// then
// pre 1
// post else
