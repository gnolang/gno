package main

func main() {
    a := len(nil)
    println(a)
}

// Error:
// main/files/len9.gno:4:10: invalid argument: nil for built-in len

// TypeCheckError:
// main/files/len9.gno:4:14: invalid argument: nil for built-in len