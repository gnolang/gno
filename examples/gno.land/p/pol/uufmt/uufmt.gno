package uufmt

import (
	"errors"
	"math"
	"strconv"
	"strings"
	"unicode"
)

const (
	flagHashtag uint8 	= 1
	flagZero    uint8	= 1 << 1
	flagMinus   uint8	= 1 << 2
	flagSpace   uint8	= 1 << 3
	flagPlus    uint8	= 1 << 4
)

const (
	lenNo uint8 = iota
	lenHH
	lenH
	lenL
	lenLL
	lenZ
)

type metadata struct {
	flags     uint8
	padding   uint
	length    uint8
	width     uint
	precision int
}

type fmtFunction func(*metadata, any) string

// formating

func formatInt(m *metadata, arg any) string {
	var str string
	switch m.length {
	case lenHH:
		str = strconv.Itoa(int(arg.(int8)))
	case lenH:
		str = strconv.Itoa(int(arg.(int16)))
	case lenL:
		str = strconv.Itoa(int(arg.(int32)))
	case lenLL:
		str = strconv.Itoa(int(arg.(int64)))
	default:
		str = strconv.Itoa(arg.(int))
	}
	if m.flags&flagSpace != 0 {
		return strings.Repeat(" ", int(math.Max(0, float64(int(m.padding)-len(str))))) + str
	} else if m.flags&flagZero != 0 {
		return strings.Repeat("0", int(math.Max(0, float64(int(m.padding)-len(str))))) + str
	}
	return str
}

func formatFloat(m *metadata, arg any) string {
	return strconv.FormatFloat(arg.(float64), 'f', m.precision, 64)
}

func formatString(m *metadata, arg any) string {
	return arg.(string)
}

func getFormatFunc(c byte) (f fmtFunction, err error) {
	var ok bool
	f, ok = map[byte]fmtFunction{
		'd': formatInt,
		'i': formatInt,
		'f': formatFloat,
		's': formatString,
	}[c]
	if !ok {
		err = errors.New("invalid rune '" + string(c) + "'")
		return
	}
	return
}

// flag parsing

func strtoi(s string) (int, string) {
	i := 0
	for i < len(s) && unicode.IsDigit(rune(s[i])) {
		i++
	}
	if i == 0 {
		return 0, s
	}
	num, err := strconv.Atoi(s[:i])
	if err != nil {
		return 0, s
	}
	return num, s[i:]
}

func (m *metadata) parse(str string, index *int) {
	*index++
	if str[*index] == '#' {
		m.flags |= flagHashtag
		*index++
	}
	c2f := map[byte]uint8 {
		'-': flagMinus,
		'+': flagPlus,
		' ': flagSpace,
		'0': flagZero,
	}
	for {
		x, ok := c2f[str[*index]]
		switch {
		case ok:
			m.flags |= x
			*index++
		case unicode.IsNumber(rune(str[*index])):
			n, rest := strtoi(str[*index:])
			if rest == str {
				continue
			}
			m.padding = uint(n)
			*index += len(strconv.Itoa(n))
		case str[*index] == '.':
			*index++
			n, rest := strtoi(str[*index:])
			if rest == str {
				continue
			}
			m.precision = n
			*index += len(strconv.Itoa(n))
		default:
			goto byebye
		}
		continue
		byebye:
		break
	}
	for k, v := range map[string]uint8 {
		"ll": lenLL,
		"l": lenL,
		"h": lenH,
		"hh": lenHH,
		"z": lenZ,
	} {
		if str[*index:*index+len(k)-1] == k {
			m.length = v
			*index += len(k)
			break
		}
	}
}

func Sprintf(base string, args ...any) string {
	largs := len(args)
	if largs == 0 {
		return base
	}
	meta := &metadata{precision: -1}
	buf := make([]byte, 0, 255)
	lbas := len(base)
	aindex := 0
	for i := 0; i < lbas; i++ {
		if base[i] == '%' && i+1 < lbas {
			meta.parse(base, &i)
			if aindex >= largs {
				return "error: too few arguments"
			}
			f, err := getFormatFunc(base[i])
			if err != nil {
				return "error:" + err.Error()
			}
			buf = append(buf, f(meta, args[aindex])...)
			i++
			aindex++
		} else {
			buf = append(buf, base[i])
		}
	}
	return string(buf)
}
