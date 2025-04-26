package main

func main() {
    a := cap(func() {})
    println(a)
}

// Error:
// main/files/cap12.gno:4:10: unexpected type for cap(): func()