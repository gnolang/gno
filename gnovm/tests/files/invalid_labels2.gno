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
// main/invalid_labels2.gno:7:3-21: cannot find branch label "FirstLoop"

// TypeCheckError:
// main/invalid_labels2.gno:7:12: invalid continue label FirstLoop; main/invalid_labels2.gno:9:1: label FirstLoop declared and not used
