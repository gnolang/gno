package main

func main() {
	var arr [5]int
	var nilArr *[5]int
	var nilSlice []int
	var nilArr2 *[8]struct{ A, B int }
	var nilMatrix *[2][3]int

	println("cap(arr):      ", cap(arr))
	println("cap(&arr):     ", cap(&arr))
	println("cap(nilArr):   ", cap(nilArr))
	println("cap(nilSlice): ", cap(nilSlice))
	println("cap(nilArr2):  ", cap(nilArr2))
	println("cap(nilMatrix):", cap(nilMatrix))

	printCap(nil)
}

func printCap(arr *[3]string) {
	println("printCap:      ", cap(arr))
}

// Output:
// cap(arr):       5
// cap(&arr):      5
// cap(nilArr):    5
// cap(nilSlice):  0
// cap(nilArr2):   8
// cap(nilMatrix): 2
// printCap:       3
