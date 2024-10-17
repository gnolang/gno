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
