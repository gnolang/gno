package ufmt

import "strconv"

// The 256-bit integer support lives in this file, behind the format256 hook,
// so that ufmt.gno names no int256 or uint256 type: `gno tool transpile`
// leaves every declaration naming them out of the Go output (Go has no
// 256-bit integer), and a printing function that named them would be left
// out with them. In Gno, init installs the hook; in transpiled Go this file's
// declarations are omitted, the hook stays nil, and only 256-bit values lose
// their formatting.

func init() {
	format256Hook = formatInt256
}

// formatInt256 returns v in decimal and its type name if v is an int256 or a
// uint256.
func formatInt256(v any) (digits, name string, ok bool) {
	switch v := v.(type) {
	case int256:
		return strconv.FormatInt256(v, 10), "int256", true
	case uint256:
		return strconv.FormatUint256(v, 10), "uint256", true
	}
	return "", "", false
}
