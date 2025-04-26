package main

func main() {
    a := make([]int, "test")
    println(a)
}

// Error:
// main/files/make5.gno:4:10: invalid length argument for built-in make: cannot convert "test" to int