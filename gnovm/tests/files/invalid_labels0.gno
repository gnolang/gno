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