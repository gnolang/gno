package main

type A nil

func main() {
}

// Error:
// main/type41.gno:3:6-11: nil is not a type

// TypeCheckError:
// main/type41.gno:3:8: nil is not a type
