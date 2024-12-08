package main

func main() {}

func undefinedLabel() {
	for i := 0; i < 10; i++ {
		break UndefinedLabel
	}
}

// Error:
// files/invalid_labels1.gno:7:9: label UndefinedLabel undefined
