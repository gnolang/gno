package main

func main() {
    a := cap(struct{}{})
    println(a)
}

// Error:
// main/files/cap13.gno:4:10: unexpected type for cap(): struct{}

// TypeCheckError:
// main/files/cap13.gno:4:14: invalid argument: struct{}{} (value of type struct{}) for built-in cap