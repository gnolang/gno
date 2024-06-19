package strconv

import "strconv"

func Itoa(n int) string                                { return strconv.Itoa(n) }
func AppendUint(dst []byte, i uint64, base int) []byte { return strconv.AppendUint(dst, i, base) }
func Atoi(s string) (int, error)                       { return strconv.Atoi(s) }
func CanBackquote(s string) bool                       { return strconv.CanBackquote(s) }
func FormatInt(i int64, base int) string               { return strconv.FormatInt(i, base) }
func FormatUint(i uint64, base int) string             { return strconv.FormatUint(i, base) }
func Quote(s string) string                            { return strconv.Quote(s) }
func QuoteToASCII(r string) string                     { return strconv.QuoteToASCII(r) }
