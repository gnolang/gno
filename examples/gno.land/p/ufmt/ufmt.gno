package ufmt

import "strconv"

func Sprintf(format string, values ...interface{}) string {
	offset := 0
	fmtlen := len(format)
	r := ""

	for i := 0; i < fmtlen; i++ {
		k := format[i]
		isLast := i == fmtlen-1
		switch {
		case k == '%' && !isLast:
			value := values[offset]
			switch format[i+1] {
			case 's':
				r += value.(string)
			case 'd':
				switch v := value.(type) {
				case int:
					r += strconv.Itoa(v)
				case int64:
					r += strconv.Itoa(int(v))
				case uint:
					r += strconv.FormatUint(uint64(v), 10)
				case uint64:
					r += strconv.FormatUint(v, 10)
				default:
					r += "(unhandled)"
				}
			case '%':
				r += "%"
			default:
				r += "(unhandled)"
			}

			i++
			offset++
		default:
			r += string(k)
		}
	}
	return r
}
