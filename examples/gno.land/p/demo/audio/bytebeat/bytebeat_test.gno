package bytebeat

import (
	"strings"
	"testing"
)

func TestByteBeat(tt *testing.T) {
	data := ByteBeat(3, 8000, func(t int) int {
		return (t>>10^t>>11)%5*((t>>14&3^t>>15&1)+1)*t%99 + ((3 + (t >> 14 & 3) - (t >> 16 & 1)) / 3 * t % 99 & 64)
	})
	if strings.Contains(data, "error") {
		tt.Fatalf("%s", data)
	}

	// TRY AT HOME: write bytes to base64
	// println(base64.StdEncoding.EncodeToString(b.Bytes()))
	// save to bytes.base64
	// using sox:
	// base64 -d bytes.base64 | play
}
