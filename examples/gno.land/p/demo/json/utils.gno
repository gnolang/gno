package json

func notDigit(c byte) bool {
	return (c & 0xF0) != 0x30
}

// lower converts a byte to lower case if it is an uppercase letter.
//
// In ASCII, the lowercase letters have the 6th bit set to 1, which is not set in their uppercase counterparts.
// This function sets the 6th bit of the byte, effectively converting uppercase letters to lowercase.
// It has no effect on bytes that are not uppercase letters.
func lower(c byte) byte {
	return c | 0x20
}

const hexLookupTable = [256]int{
	'0': 0x0, '1': 0x1, '2': 0x2, '3': 0x3, '4': 0x4,
	'5': 0x5, '6': 0x6, '7': 0x7, '8': 0x8, '9': 0x9,
	'A': 0xA, 'B': 0xB, 'C': 0xC, 'D': 0xD, 'E': 0xE, 'F': 0xF,
	'a': 0xA, 'b': 0xB, 'c': 0xC, 'd': 0xD, 'e': 0xE, 'f': 0xF,
	// Fill unspecified index-value pairs with key and value of -1
	'G': -1, 'H': -1, 'I': -1, 'J': -1,
	'K': -1, 'L': -1, 'M': -1, 'N': -1,
	'O': -1, 'P': -1, 'Q': -1, 'R': -1,
	'S': -1, 'T': -1, 'U': -1, 'V': -1,
	'W': -1, 'X': -1, 'Y': -1, 'Z': -1,
	'g': -1, 'h': -1, 'i': -1, 'j': -1,
	'k': -1, 'l': -1, 'm': -1, 'n': -1,
	'o': -1, 'p': -1, 'q': -1, 'r': -1,
	's': -1, 't': -1, 'u': -1, 'v': -1,
	'w': -1, 'x': -1, 'y': -1, 'z': -1,
}

func h2i(c byte) int {
	return hexLookupTable[c]
}

func trimNegativeSign(bytes []byte) (neg bool, trimmed []byte) {
	if bytes[0] == '-' {
		return true, bytes[1:]
	}

	return false, bytes
}

/*		Testing Helper Function		*/

func isEqualMap(a, b map[string]interface{}) bool {
	if len(a) != len(b) {
		return false
	}

	for key, valueA := range a {
		valueB, ok := b[key]
		if !ok {
			return false
		}

		switch valueA := valueA.(type) {
		case []interface{}:
			if valueB, ok := valueB.([]interface{}); ok {
				if !isEqualSlice(valueA, valueB) {
					return false
				}
			} else {
				return false
			}
		default:
			if valueA != valueB {
				return false
			}
		}
	}

	return true
}

func isEqualSlice(a, b []interface{}) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func fieldsMatch(fs []Field, names []string) bool {
	if len(fs) != len(names) {
		return false
	}

	for i, f := range fs {
		if f.name != names[i] {
			return false
		}
	}

	return true
}

func fieldNames(fs []Field) []string {
	names := make([]string, len(fs))
	for i, f := range fs {
		names[i] = f.name
	}

	return names
}
