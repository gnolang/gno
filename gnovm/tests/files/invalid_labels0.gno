package main

func main() {}
func invalidLabel() {
FirstLoop:
	for i := 0; i < 10; i++ {
	}
	for i := 0; i < 10; i++ {
		break FirstLoop
	}
}

// Error:
// main/invalid_labels0.gno:9:3-18: cannot find branch label "FirstLoop"

// TypeCheckError:
// main/invalid_labels0.gno:9:9: invalid break label FirstLoop; main/invalid_labels0.gno:5:1: label FirstLoop declared and not used
