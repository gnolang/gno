package main

import bar "filetests/extern/foo_bar"

func main() {
	println(bar.Name)
}

// Output:
// foo_bar
