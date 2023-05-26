// PKGPATH: gno.land/r/demo/tests_test
package tests_test

var (
	ss1 = []string{"A1"}
	ss2 = []string{"A2"}
)

// this bug seems to be limited to slices and potentially only when using `println`.

func main() {
	println("ss1", ss1) // unexpected print
	println("ss1", ss1) // unexpected print
	_ = ss1
	println("ss1", ss1) // expected print

	println("ss2", ss2, ss2[0]) // expected print
}

// Output:
// ss1 slice[ref(0e4830f6b0cb49c615cab8df855d8dd0f7ebc58c:4)]
// ss1 slice[ref(0e4830f6b0cb49c615cab8df855d8dd0f7ebc58c:4)]
// ss1 slice[("A1" string)]
// ss2 slice[("A2" string)] A2
