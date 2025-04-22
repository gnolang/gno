package main

func main() {
    for i := 0; i < 10; i++ {
        if i == 1 {
            _ = func() int {
                fallthrough
                return 11
            }()
        }
        println(i)
    }
    println("wat???")
}

// Error:
// main/files/for22.gno:7:17: fallthrough statement out of place

// TypeCheckError:
// main/files/for22.gno:7:17: fallthrough statement out of place
