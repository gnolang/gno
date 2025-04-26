package main

func main() {}

func invalidLabelStatement() {
	if true {
		break InvalidLabel
	}
}

// Error:
// files/invalid_labels3.gno:7:9: label InvalidLabel undefined

// TypeCheckError:
// main/files/invalid_labels3.gno:7:9: invalid break label InvalidLabel
