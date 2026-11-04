package disperse

import (
	"math"
	"testing"

	"gno.land/p/nt/uassert/v0"
)

func TestParseTokens(t *testing.T) {
	amounts, symbols, err := parseTokens("150TEST,115792089237316195423570985008687907853269984665640564039457584007913129639935BIG,0Z")
	uassert.NoError(t, err)
	uassert.Equal(t, 3, len(amounts))
	equal256(t, uint256(150), amounts[0])
	equal256(t, uint256(math.MaxUint256), amounts[1])
	equal256(t, uint256(0), amounts[2])
	uassert.Equal(t, "TEST", symbols[0])
	uassert.Equal(t, "BIG", symbols[1])
	uassert.Equal(t, "Z", symbols[2])
}

func TestParseTokensRejectsMalformed(t *testing.T) {
	for _, in := range []string{
		// One past MaxUint256: refused, not wrapped.
		"115792089237316195423570985008687907853269984665640564039457584007913129639936BIG",
		"-1BIG",  // no sign
		"+1BIG",  // no sign
		"BIG",    // no amount
		"",       // nothing at all
		"10,5X",  // an entry without a symbol has no amount either
		"5X,,6Y", // an empty entry
	} {
		_, _, err := parseTokens(in)
		uassert.ErrorIs(t, err, ErrInvalidAmount, "input "+in)
	}
}
