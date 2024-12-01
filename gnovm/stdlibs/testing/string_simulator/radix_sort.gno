package main

func counting_sort(arr []int, exp int) []int {
	n := len(arr)
	output := make([]int, n)
	count := make([]int, 10) // 0~9 자릿수 카운트

	// 자릿수 값 카운트
	for i := 0; i < n; i++ {
		index := (arr[i] / exp) % 10
		count[index]++
	}

	// 누적 카운트 계산
	for i := 1; i < 10; i++ {
		count[i] += count[i-1]
	}

	// 정렬 결과 저장
	for i := n - 1; i >= 0; i-- {
		index := (arr[i] / exp) % 10
		output[count[index]-1] = arr[i]
		count[index]--
	}

	// 결과를 원본 배열에 복사
	for i := 0; i < n; i++ {
		arr[i] = output[i]
	}

	return arr
}

func radixSort(arr []int) {
	// 최대값 찾기
	max := 0
	for _, v := range arr {
		if v > max {
			max = v
		}
	}

	// 자릿수별 정렬
	for exp := 1; max/exp > 0; exp *= 10 {
		arr = counting_sort(arr, exp)
	}
}

func main() {
	// 예제: 10만 개의 숫자 생성
	n := 100000
	arr := make([]int, n)
	for i := 0; i < n; i++ {
		arr[i] = (i % 10000) + 1 // 1~10,000 범위의 숫자
	}

	println("Sorting...")
	radixSort(arr)
	println("Sorted!")
}
