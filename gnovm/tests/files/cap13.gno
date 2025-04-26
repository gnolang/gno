package main

func main() {
    a := cap(struct{}{})
    println(a)
}

// Error:
// main/files/cap13.gno:4:10: unexpected type for cap(): struct{}