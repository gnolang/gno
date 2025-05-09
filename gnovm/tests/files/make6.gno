package main

func main() {
    a := make([]int, 10, "5")
    println(a)
}

// Error:
// main/files/make6.gno:4:10: invalid capacity argument for built-in make: cannot convert "5" to int

// TypeCheckError:
// main/files/make6.gno:4:26: cannot convert "5" (untyped string constant) to type int