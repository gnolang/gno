package main

func main() {
    a := cap(nil)
    println(a)
}

// Error:
// main/files/cap11.gno:4:10: invalid argument: nil for built-in cap

// TypeCheckError:
// main/files/cap11.gno:4:14: invalid argument: nil for built-in cap