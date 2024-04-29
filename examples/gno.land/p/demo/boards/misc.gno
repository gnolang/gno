package boards

import (
	"strconv"
	"strings"
)

func postIDKey(pid PostID) string {
	return padZero(uint64(pid), 10)
}

func padZero(u64 uint64, length int) string {
	str := strconv.Itoa(int(u64))
	if len(str) >= length {
		return str
	} else {
		return strings.Repeat("0", length-len(str)) + str
	}
}

// NOTE: length must be greater than 3.
func summaryOf(str string, length int) string {
	lines := strings.SplitN(str, "\n", 2)
	line := lines[0]
	if len(line) > length {
		line = line[:(length-3)] + "..."
	} else if len(lines) > 1 {
		// len(line) <= 80
		line = line + "..."
	}
	return line
}
