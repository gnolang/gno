package json

func notDigit(c byte) bool {
	return c < '0' || c > '9'
}

func h2i(c byte) int {
	switch {
	case c >= '0' && c <= '9':
		return int(c - '0')
	case c >= 'A' && c <= 'F':
		return int(c - 'A' + 10)
	case c >= 'a' && c <= 'f':
		return int(c - 'a' + 10)
	}

	return BadHex
}

func trimNegativeSign(bytes []byte) (neg bool, trimmed []byte) {
	if bytes[0] == '-' {
		return true, bytes[1:]
	}

	return false, bytes
}
