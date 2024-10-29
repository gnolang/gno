package main

func main() {
    for i := 0; i < 10; i++ {
        if i == 1 {
            _ = func() int {
                break
                return 11
            }()
        }
        println(i)
    }
    println("wat???")
}

// Error:
// break statement out of place