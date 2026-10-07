// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package url

// Custom code: Go validates the contents of an IP-literal with
// netip.ParseAddr. Gno has no net/netip (it is non-deterministic host
// networking), so this file provides the small amount of address parsing
// that net/url actually needs: deciding whether the text between "[" and
// "]" is a valid IPv6 address.
//
// The grammar implemented here mirrors net/netip's IPv6 parser: up to eight
// groups of 1-4 hex digits separated by ':', at most one "::" run-of-zeros
// marker, an optional trailing dotted-quad for IPv4-mapped forms, and an
// optional "%zone" suffix.

// isIPv6Literal reports whether s is a valid IPv6 address, optionally
// followed by a "%zone" suffix. IPv4 addresses are not accepted, matching
// Go's rule that only an IPv6 address may be enclosed in square brackets.
func isIPv6Literal(s string) bool {
	// Split off the zone, if any. A zone must be non-empty.
	for i := 0; i < len(s); i++ {
		if s[i] == '%' {
			if i+1 >= len(s) {
				return false
			}
			zone := s[i+1:]
			for j := 0; j < len(zone); j++ {
				// Keep zones boring: no control characters or separators.
				if zone[j] < ' ' || zone[j] == 0x7f || zone[j] == '/' || zone[j] == '[' || zone[j] == ']' {
					return false
				}
			}
			s = s[:i]
			break
		}
	}
	return parseIPv6(s)
}

// parseIPv6 reports whether s is a valid IPv6 address with no zone.
func parseIPv6(s string) bool {
	// ellipsis is the byte index of the start of the "::" run, or -1.
	ellipsis := -1

	if len(s) >= 2 && s[0] == ':' && s[1] == ':' {
		ellipsis = 0
		s = s[2:]
		if len(s) == 0 {
			return true // "::"
		}
	}

	// Loop, parsing hex numbers followed by colon.
	filled := 0 // number of 16-bit groups filled
	for filled < 8 {
		// Hex number: 1-4 digits.
		n := 0
		digits := 0
		for n < len(s) {
			d := hexDigit(s[n])
			if d < 0 {
				break
			}
			digits++
			if digits > 4 {
				return false
			}
			n++
		}
		if digits == 0 {
			return false
		}

		// If followed by dot, might be an embedded IPv4 address.
		if n < len(s) && s[n] == '.' {
			if ellipsis < 0 && filled != 6 {
				return false // not enough room
			}
			if filled+2 > 8 {
				return false
			}
			if !parseIPv4(s) {
				return false
			}
			s = ""
			filled += 2
			break
		}

		// Save this 16-bit chunk.
		s = s[n:]
		filled++

		// Stop at end of string.
		if len(s) == 0 {
			break
		}

		// Otherwise must be followed by colon and more.
		if s[0] != ':' || len(s) == 1 {
			return false
		}
		s = s[1:]

		// Look for ellipsis.
		if s[0] == ':' {
			if ellipsis >= 0 { // already have one
				return false
			}
			ellipsis = filled
			s = s[1:]
			if len(s) == 0 { // can be at end
				break
			}
		}
	}

	// Must have used entire string.
	if len(s) != 0 {
		return false
	}

	// If didn't parse enough, expand ellipsis.
	if filled < 8 {
		if ellipsis < 0 {
			return false
		}
	} else if ellipsis >= 0 {
		// Ellipsis must represent at least one 0 group.
		return false
	}
	return true
}

// parseIPv4 reports whether s is a valid dotted-quad IPv4 address.
func parseIPv4(s string) bool {
	for i := 0; i < 4; i++ {
		if len(s) == 0 {
			return false
		}
		if i > 0 {
			if s[0] != '.' {
				return false
			}
			s = s[1:]
		}
		n, c, ok := dtoi(s)
		if !ok || n > 0xFF {
			return false
		}
		if c > 1 && s[0] == '0' {
			return false // reject leading zeros
		}
		s = s[c:]
	}
	return len(s) == 0
}

// dtoi returns the leading non-negative decimal integer in s,
// the number of bytes consumed, and whether it parsed at all.
func dtoi(s string) (n int, i int, ok bool) {
	const big = 0xFFFFFF
	for i = 0; i < len(s) && '0' <= s[i] && s[i] <= '9'; i++ {
		n = n*10 + int(s[i]-'0')
		if n >= big {
			return big, i, false
		}
	}
	if i == 0 {
		return 0, 0, false
	}
	return n, i, true
}

// hexDigit returns the value of the hex digit b, or -1.
func hexDigit(b byte) int {
	switch {
	case '0' <= b && b <= '9':
		return int(b - '0')
	case 'a' <= b && b <= 'f':
		return int(b-'a') + 10
	case 'A' <= b && b <= 'F':
		return int(b-'A') + 10
	}
	return -1
}
