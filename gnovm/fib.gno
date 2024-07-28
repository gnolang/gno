// fib.gno
package main

// fib returns the nth number in the Fibonacci sequence.
func fib(n int) int {
        if n < 2 {
                return n
        }
        return fib(n-2) + fib(n-1)
}

func main() {
        println(fib(4))
}
