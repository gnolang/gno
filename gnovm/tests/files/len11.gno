package main

func main() {
    a := len(func() {})
    println(a)
}

// Error:
// main/files/len11.gno:4:10: unexpected type for len(): func()

// TypeCheckError:
// main/files/len11.gno:4:14: invalid argument: (func() literal) (value of type func()) for built-in len