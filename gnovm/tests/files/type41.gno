package main

type A nil

func main() {
}

// Error:
// main/files/type41.gno:3:6: nil is not a type

// TypeCheckError:
// main/files/type41.gno:3:8: nil is not a type
