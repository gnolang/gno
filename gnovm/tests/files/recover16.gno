package main


func main() {
    defer func() {
    		r := recover()
    		println("recover:", r)
    }()

    _ = make([]int, -1) // Panics because of negative length
}

// Output:
// recover: len out of range
