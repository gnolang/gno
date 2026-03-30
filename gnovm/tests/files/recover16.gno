package main


func main() {
    defer func() {
    		r := recover()
    		println("recover:", r)
    }()

    _ = make([]int, -1) // Panics because of negative length
}

// TypeCheckError:
// main/recover16.gno:10:21: invalid argument: index -1 (constant of type int) must not be negative

// Error:
// main/recover16.gno:10:9-24: invalid argument: index (-1 <untyped> bigint) must not be negative
