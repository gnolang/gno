package main

// These imports require an identifier, as the `package` declaration doesn't
// match their paths.

import (
	tutu "filetests/extern/tata"
	titi "filetests/extern/toto"
)

func main() {
	println("Hello", titi.Quux())
	println("Hello", tutu.Quux())
}

// Output:
// Hello bar
// Hello baz
