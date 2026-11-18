package gnolang

import (
	"math/big"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A short string must not decode into a much larger number: this decode runs
// on untrusted transaction bytes before any signature or fee check.
func TestBigdecUnmarshalAminoRefusesExpansion(t *testing.T) {
	t.Parallel()

	for _, s := range []string{
		// Decimal exponents, which expand to the exponent's digit count.
		"1e999999", "1E999999", "1e-999999", "1.5e999999", "1e+999999",
		// Binary exponents, which SetString bounds far less tightly.
		"1p999999", "1P999999", "1p-999999", "1p9999999",
		// The same exponents behind a base prefix.
		"0x1p999999", "0b1p999999", "0o1p999999",
		// Other spellings MarshalAmino never produces.
		"1.5", "0x10", "1_000", "+5", " 1", "1 ",
		// The right characters in the wrong shape.
		"", "-", "/", "1/", "/3", "1/0", "1/-3", "1-2", "1/2/3", "--1",
	} {
		var bdv BigdecValue
		assert.Error(t, bdv.UnmarshalAmino(s), "%q", s)
		assert.Nil(t, bdv.V, "%q", s)
	}
}

// Everything MarshalAmino emits must keep decoding: realm state and transaction
// replay go through this same decode.
func TestBigdecAminoRoundTrip(t *testing.T) {
	t.Parallel()

	huge := new(big.Int).Exp(big.NewInt(10), big.NewInt(5000), nil)
	for _, r := range []*big.Rat{
		new(big.Rat),
		big.NewRat(1, 1),
		big.NewRat(-5, 1),
		big.NewRat(1, 3),
		big.NewRat(-22, 7),
		new(big.Rat).SetInt(huge),
		new(big.Rat).SetFrac(big.NewInt(1), huge),
	} {
		s, err := BigdecValue{V: r}.MarshalAmino()
		require.NoError(t, err)

		var got BigdecValue
		require.NoError(t, got.UnmarshalAmino(s), "%q", s)
		assert.Zero(t, r.Cmp(got.V), "%q", s)
	}

	// A nil value marshals to "0".
	s, err := BigdecValue{}.MarshalAmino()
	require.NoError(t, err)
	var zero BigdecValue
	require.NoError(t, zero.UnmarshalAmino(s))
	assert.Zero(t, zero.V.Sign())

	// The float form is exact, including an exponent too large for a rat.
	f, _, err := big.ParseFloat("1e999999", 0, BigdecFloatPrec, big.ToNearestEven)
	require.NoError(t, err)
	s, err = BigdecValue{F: f}.MarshalAmino()
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(s, bigdecFloatMarshalPrefix), s)
	var fl BigdecValue
	require.NoError(t, fl.UnmarshalAmino(s))
	assert.Zero(t, f.Cmp(fl.F))
}
