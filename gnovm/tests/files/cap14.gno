package main

func main() {
    a := cap(map[string]int{})
    println(a)
}

// Error:
// main/files/cap14.gno:4:10: unexpected type for cap(): map[string]int