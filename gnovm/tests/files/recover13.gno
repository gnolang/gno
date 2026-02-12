package main


func main() {
    defer func() {
    		r := recover()
    		println("recover:", r)
    }()

    arr := []int{1, 2, 3}
    _ = arr[-1:] // Panics because of negative index
}

// TypeCheckError:
// main/recover13.gno:11:13: invalid argument: index -1 (constant of type int) must not be negative

// Error:
// main/recover13.gno:11:9-17: invalid argument, index must not be negative: (-1 int)
