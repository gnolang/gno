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

// TypeCheckError:
// main/files/recover16.gno:10:21: invalid argument: index -1 (constant of type int) must not be negative
