package main

func main() {
    a := make(func() {})
    println(a)
}

// Error:
// main/files/make4.gno:4:10: func func(){  } is not a type

// TypeCheckError:
// main/files/make4.gno:4:15: (func() literal) is not a type