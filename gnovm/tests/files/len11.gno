package main

func main() {
    a := len(func() {})
    println(a)
}

// Error:
// main/files/len11.gno:4:10: unexpected type for len(): func()