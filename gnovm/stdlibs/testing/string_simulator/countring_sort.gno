package main

func countingSort(arr []int, maxVal int) []int {
	count := make([]int, maxVal+1) // 0 ~ maxVal
	output := make([]int, len(arr))

	// 빈도 계산
	for _, num := range arr {
		count[num]++
	}

	// 누적 합 계산
	for i := 1; i <= maxVal; i++ {
		count[i] += count[i-1]
	}

	// 정렬된 결과 생성
	for i := len(arr) - 1; i >= 0; i-- {
		output[count[arr[i]]-1] = arr[i]
		count[arr[i]]--
	}

	return output
}

// func main() {
// 	// 예제 데이터
// 	arr := []int{100000, 99999, 1, 50000, 25000, 75000, 100000}
// 	sorted := countingSort(arr, 100000)

// 	println("Sorted array:", sorted)
// }
