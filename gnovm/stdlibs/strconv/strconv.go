package strconv

import "strconv"

func Itoa(n int) string                                { return strconv.Itoa(n) }
func AppendUint(dst []byte, i uint64, base int) []byte { return strconv.AppendUint(dst, i, base) }
func Atoi(s string) (int, error)                       { return strconv.Atoi(s) }
func CanBackquote(s string) bool                       { return strconv.CanBackquote(s) }
func FormatInt(i int64, base int) string               { return strconv.FormatInt(i, base) }
func FormatUint(i uint64, base int) string             { return strconv.FormatUint(i, base) }
func FormatFloat(f float64, fmt byte, prec, bitSize int) string {
	return strconv.FormatFloat(f, fmt, prec, bitSize)
}
func Quote(s string) string        { return strconv.Quote(s) }
func QuoteToASCII(r string) string { return strconv.QuoteToASCII(r) }
