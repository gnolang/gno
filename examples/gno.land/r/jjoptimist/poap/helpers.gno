package poap

func padRight(s string, total int) string {
	n := total - len([]rune(s))
	if n < 0 {
		n = 0
	}
	padding := ""
	for i := 0; i < n; i++ {
		padding += " "
	}
	return s + padding
}

func repeat(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}

func centerText(s string, total int) string {
	runes := []rune(s)
	l := len(runes)
	if l >= total {
		return s
	}
	spaces := total - l
	leftSpaces := spaces / 2
	rightSpaces := spaces - leftSpaces
	return repeat(" ", leftSpaces) + s + repeat(" ", rightSpaces)
} 