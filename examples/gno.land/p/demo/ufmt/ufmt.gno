package ufmt

import "strconv"

func Sprintf(format string, args ...interface{}) string {
	end := len(format)
	argNum := 0
	argLen := len(args)
	buf := ""

	for i := 0; i < end; {
		isLast := i == end-1
		c := format[i]

		if isLast || c != '%' {
			// we don't check for invalid format like a one ending with "%"
			buf += string(c)
			i++
			continue
		}

		verb := format[i+1]
		if verb == '%' {
			buf += "%"
			i += 2
			continue
		}

		if argNum > argLen {
			panic("invalid number of arguments to ufmt.Sprintf")
		}
		arg := args[argNum]
		argNum++

		switch verb {
		case 's':
			buf += arg.(string)
		case 'd':
			switch v := arg.(type) {
			case int:
				buf += strconv.Itoa(v)
			case int64:
				buf += strconv.Itoa(int(v))
			case uint:
				buf += strconv.FormatUint(uint64(v), 10)
			case uint64:
				buf += strconv.FormatUint(v, 10)
			default:
				buf += "(unhandled)"
			}
		case 't':
			switch v := arg.(type) {
			case bool:
				if v {
					buf += "true"
				} else {
					buf += "false"
				}
			default:
				buf += "(unhandled)"
			}
		case '%':
			buf += "%"
		default:
			buf += "(unhandled)"
		}

		i += 2
	}
	if argNum < argLen {
		panic("too many arguments to ufmt.Sprintf")
	}
	return buf
}
