package strconv

func Itoa(n int) string                                // injected
func AppendUint(dst []byte, i uint64, base int) []byte // injected
func Atoi(s string) (int, error)                       // injected
func CanBackquote(s string) bool                       // injected
func FormatInt(i int64, base int) string               // injected
func FormatUint(i uint64, base int) string             // injected
func Quote(s string) string                            // injected
func QuoteToASCII(s string) string                     // injected
