package main

func main() {
	var byteArray []byte = []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
	arrayTwenty := [20]byte(byteArray)
	println(arrayTwenty)
}

// Output:
// array[0x0102030405060708090A0B0C0D0E0F1011121314]
