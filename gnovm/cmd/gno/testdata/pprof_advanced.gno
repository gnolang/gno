package main

// Simulate a web request handler with multiple layers

func main() {
	// Simulate handling multiple requests
	for i := 0; i < 5; i++ {
		handleRequest(i)
	}
	println("Done processing requests")
}

func handleRequest(id int) {
	// Validate input
	if !validateInput(id) {
		return
	}

	// Process the request
	data := fetchData(id)
	result := processData(data)
	storeResult(id, result)
}

func validateInput(id int) bool {
	// Simulate validation logic
	return id >= 0 && id < 100
}

func fetchData(id int) []int {
	// Simulate fetching data from storage
	data := make([]int, 20)
	for i := 0; i < 20; i++ {
		data[i] = computeValue(id, i)
	}
	return data
}

func computeValue(id, index int) int {
	// Heavy computation
	result := id * index
	for j := 0; j < 10; j++ {
		result = result + j*j
	}
	return result
}

func processData(data []int) int {
	// Aggregate the data
	total := 0
	for _, v := range data {
		total += transform(v)
	}
	return total
}

func transform(value int) int {
	// Apply transformation
	if value < 0 {
		value = -value
	}
	return value * 2 + 1
}

func storeResult(id int, result int) {
	// Simulate storing result
	cache := make(map[string]int)
	key := "result_" + intToString(id)
	cache[key] = result
}

func intToString(n int) string {
	if n == 0 {
		return "0"
	}

	neg := n < 0
	if neg {
		n = -n
	}

	// Build string digit by digit
	digits := make([]byte, 0, 10)
	for n > 0 {
		digits = append(digits, byte('0'+n%10))
		n = n / 10
	}

	// Reverse
	for i, j := 0, len(digits)-1; i < j; i, j = i+1, j-1 {
		digits[i], digits[j] = digits[j], digits[i]
	}

	if neg {
		return "-" + string(digits)
	}
	return string(digits)
}
