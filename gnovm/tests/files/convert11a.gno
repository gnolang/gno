package main

func main() {

	var nilSlice []int
	arrPtr := (*[3]int)(nilSlice) // Panic: runtime error: cannot convert slice with length 0 to array or pointer to array with length 3
	println(len(arrPtr))

}

// Error:
// runtime error: cannot convert slice with length 0 to array or pointer to array with length 3
