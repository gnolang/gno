package main

func main() {
	const t = [0]string{}
}

// Error:
// main/const40.gno:4:8-23: (const-type [0]string){} (variable of type [0]string) is not constant

// TypeCheckError:
// main/const40.gno:4:12: [0]string{} (value of type [0]string) is not constant
