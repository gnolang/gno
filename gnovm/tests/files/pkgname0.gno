package main

import (
	quux "filetests/extern/bar" // pkg name is actually quux
	baz "filetests/extern/baz"  // pkg name is also quux, force it to be baz.
)

func main() {
	println("Hello", quux.Quux())
	println("Hello", baz.Quux())
}

// Output:
// Hello bar
// Hello baz
