package main

func fibonacci(n int) int {
	if n <= 1 {
		return n
	}
	return fibonacci(n-1) + fibonacci(n-2)
}

func sum(arr []int) int {
	total := 0
	for _, v := range arr {
		total += v
	}
	return total
}

func buildArray(size int) []int {
	arr := make([]int, size)
	for i := 0; i < size; i++ {
		arr[i] = i * i
	}
	return arr
}

func main() {
	// Test recursion
	fib := fibonacci(15)
	println("fibonacci(15) =", fib)

	// Test loops and arrays
	arr := buildArray(100)
	s := sum(arr)
	println("sum of squares =", s)

	// Test map operations
	m := make(map[string]int)
	for i := 0; i < 50; i++ {
		key := "key" + string(rune('0'+i%10))
		m[key] = i
	}
	println("map size =", len(m))
}
