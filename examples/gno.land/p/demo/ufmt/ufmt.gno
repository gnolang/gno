// Package ufmt provides utility functions for formatting strings, similarly
// to the Go package "fmt", of which only a subset is currently supported
// (hence the name µfmt - micro fmt).
package ufmt

import (
	"strconv"
	"strings"
)

// Println formats using the default formats for its operands and writes to standard output.
// Println writes the given arguments to standard output with spaces between arguments
// and a newline at the end.
func Println(args ...interface{}) {
	var strs []string
	for _, arg := range args {
		switch v := arg.(type) {
		case string:
			strs = append(strs, v)
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			strs = append(strs, Sprintf("%d", v))
		case bool:
			if v {
				strs = append(strs, "true")

				continue
			}

			strs = append(strs, "false")
		default:
			strs = append(strs, "(unhandled)")
		}
	}

	// TODO: remove println after gno supports os.Stdout
	println(strings.Join(strs, " "))
}

// Sprintf offers similar functionality to Go's fmt.Sprintf, or the sprintf
// equivalent available in many languages, including C/C++.
// The number of args passed must exactly match the arguments consumed by the format.
// A limited number of formatting verbs and features are currently supported,
// hence the name ufmt (µfmt, micro-fmt).
//
// The currently formatted verbs are the following:
//
//	%s: places a string value directly.
//	    If the value implements the interface interface{ String() string },
//	    the String() method is called to retrieve the value.
//	%d: formats an integer value using package "strconv".
//	    Currently supports only uint, uint64, int, int64.
//	%t: formats a boolean value to "true" or "false".
//	%x: formats an integer value as a hexadecimal string.
//	    Currently supports only uint8, []uint8, [32]uint8.
//	%c: formats a rune value as a string.
//	    Currently supports only rune, int.
//	%q: formats a string value as a quoted string.
//	%T: formats the type of the value.
//	%%: outputs a literal %. Does not consume an argument.
func Sprintf(format string, args ...interface{}) string {
	// we use runes to handle multi-byte characters
	sTor := []rune(format)
	end := len(sTor)
	argNum := 0
	argLen := len(args)
	buf := ""

	for i := 0; i < end; {
		isLast := i == end-1
		c := string(sTor[i])

		if isLast || c != "%" {
			// we don't check for invalid format like a one ending with "%"
			buf += string(c)
			i++
			continue
		}

		verb := string(sTor[i+1])
		if verb == "%" {
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
		case "s":
			switch v := arg.(type) {
			case interface{ String() string }:
				buf += v.String()
			case string:
				buf += v
			default:
				buf += "(unhandled)"
			}
		case "d":
			switch v := arg.(type) {
			case int:
				buf += strconv.Itoa(v)
			case int8:
				buf += strconv.Itoa(int(v))
			case int16:
				buf += strconv.Itoa(int(v))
			case int32:
				buf += strconv.Itoa(int(v))
			case int64:
				buf += strconv.Itoa(int(v))
			case uint:
				buf += strconv.FormatUint(uint64(v), 10)
			case uint8:
				buf += strconv.FormatUint(uint64(v), 10)
			case uint16:
				buf += strconv.FormatUint(uint64(v), 10)
			case uint32:
				buf += strconv.FormatUint(uint64(v), 10)
			case uint64:
				buf += strconv.FormatUint(v, 10)
			default:
				buf += "(unhandled)"
			}
		case "t":
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
		case "x":
			switch v := arg.(type) {
			case uint8:
				buf += strconv.FormatUint(uint64(v), 16)
			case []uint8, [32]uint8:
				for _, b := range v {
					buf += strconv.FormatUint(uint64(b), 16)
				}
			default:
				buf += "(unhandled)"
			}
		case "c":
			switch v := arg.(type) {
			case rune:
				buf += string(v)
			case int:
				buf += string(rune(v))
			default:
				buf += "(unhandled)"
			}
		case "q":
			switch v := arg.(type) {
			case string:
				buf += strconv.Quote(v)
			default:
				buf += "(unhandled)"
			}
		case "T":
			switch arg.(type) {
			case bool:
				buf += "bool"
			case int:
				buf += "int"
			case int8:
				buf += "int8"
			case int16:
				buf += "int16"
			case int32:
				buf += "int32"
			case int64:
				buf += "int64"
			case uint:
				buf += "uint"
			case uint8:
				buf += "uint8"
			case uint16:
				buf += "uint16"
			case uint32:
				buf += "uint32"
			case uint64:
				buf += "uint64"
			case string:
				buf += "string"
			case []byte:
				buf += "[]byte"
			case []rune:
				buf += "[]rune"
			default:
				buf += "unknown"
			}
		// % handled before, as it does not consume an argument
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

// errMsg implements the error interface.
type errMsg struct {
	msg string
}

// Error defines the requirements of the error interface.
// It functions similarly to Go's errors.New()
func (e *errMsg) Error() string {
	return e.msg
}

// Errorf is a function that mirrors the functionality of fmt.Errorf.
//
// It takes a format string and arguments to create a formatted string,
// then sets this string as the 'msg' field of an errMsg struct and returns a pointer to this struct.
//
// This function operates in a similar manner to Go's fmt.Errorf,
// providing a way to create formatted error messages.
//
// The currently formatted verbs are the following:
//
//	%s: places a string value directly.
//	    If the value implements the interface interface{ String() string },
//	    the String() method is called to retrieve the value.
//	%d: formats an integer value using package "strconv".
//	    Currently supports only uint, uint64, int, int64.
//	%t: formats a boolean value to "true" or "false".
//	%%: outputs a literal %. Does not consume an argument.
func Errorf(format string, args ...interface{}) error {
	return &errMsg{Sprintf(format, args...)}
}
