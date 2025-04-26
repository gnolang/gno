package main

func main() {}

func labelOutsideScope() {
	for i := 0; i < 10; i++ {
		continue FirstLoop
	}
FirstLoop:
	for i := 0; i < 10; i++ {
	}
}

// Error:
// main/files/invalid_labels2.gno:7:3: cannot find branch label "FirstLoop"

// TypeCheckError:
// main/files/invalid_labels2.gno:7:12: invalid continue label FirstLoop; main/files/invalid_labels2.gno:9:1: label FirstLoop declared and not used
