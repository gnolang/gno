package main

func main() {
	expected := `hello`[:] // this kind of untyped non const will be converted

	a := 1

	println(a == expected)
}

// Error:
// main/files/types/0a1f_filetest.gno:8: cannot use int as string
