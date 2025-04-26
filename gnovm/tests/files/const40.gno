package main

func main() {
	const t = [0]string{}
}

// Error:
// main/files/const40.gno:4:8: [(const (0 int))](const-type string){} (variable of type [0]string) is not constant

// TypeCheckError:
// main/files/const40.gno:4:12: [0]string{} (value of type [0]string) is not constant
