package example

func Add(a, b int) int {
	if a < 0 || b < 0 {
		return 0
	}
	return a + b
}

func Multiply(a, b int) int {
	if a == 0 || b == 0 {
		return 0
	}
	return a * b
}