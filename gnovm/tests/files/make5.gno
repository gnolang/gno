package main

func main() {
    a := make([]int, "test")
    println(a)
}

// Error:
// main/files/make5.gno:4:10: invalid length argument for built-in make: cannot convert "test" to int

// TypeCheckError:
// main/files/make5.gno:4:22: cannot convert "test" (untyped string constant) to type int