package strconv

// All injected

func Itoa(n int) string
func AppendUint(dst []byte, i uint64, base int) []byte
func Atoi(s string) (int, error)
func CanBackquote(s string) bool
func FormatInt(i int64, base int) string
func FormatUint(i uint64, base int) string
func Quote(s string) string
func QuoteToASCII(s string) string
