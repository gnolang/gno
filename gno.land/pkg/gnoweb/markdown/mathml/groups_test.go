package mathml

import (
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// convertErrWithin converts tex inline, failing the test if conversion
// does not finish within two seconds, and returns the conversion error.
func convertErrWithin(t *testing.T, tex string) (string, error) {
	t.Helper()
	type result struct {
		out string
		err error
	}
	done := make(chan result, 1)
	go func() {
		out, err := NewMathMLConverter().ConvertInline(tex)
		done <- result{out, err}
	}()
	select {
	case r := <-done:
		return r.out, r.err
	case <-time.After(2 * time.Second):
		t.Fatalf("conversion of %q did not terminate", tex)
		return "", nil
	}
}

// A {group} that opens inside an environment or a \left...\right pair and
// closes outside it, or the reverse, is an error in TeX ("Extra }"). The
// converter rejects it too, so the expression falls back to its source,
// instead of reading a group past the end of the one around it.
func TestCrossingGroupIsRejected(t *testing.T) {
	for _, tex := range []string{
		`\left(\color x{\right)}`,
		`\begin{matrix}\color x{a\end{matrix}}`,
		`{\left( x}\right)`,
		`\left( {x \right)}`,
		`{\begin{matrix} a}\end{matrix}`,
		`\begin{matrix}{a\end{matrix}}`,
		`\left( \begin{matrix} a \right) \end{matrix}`,
		`\begin{matrix} \left( a \end{matrix} \right)`,
		`{\left(` + `x` + `}\right)`,
		`{\left({\left(x}\right)}\right)`,
	} {
		_, err := convertErrWithin(t, tex)
		assert.Error(t, err, tex)
	}
}

// Pairs of plain brackets are not groups in TeX: one may straddle a group,
// and the expression still converts.
func TestPlainBracketsMayStraddleGroups(t *testing.T) {
	for tex, want := range map[string]string{
		`({)}`:               `<mo form="prefix" stretchy="false">(</mo>`,
		`x^{(} y)`:           `<mi>y</mi>`,
		`\sqrt[{]}x`:         `<mi>x</mi>`,
		`[{a)}`:              `<mi>a</mi>`,
		`\left( a \right)`:   `<mi>a</mi>`,
		`\left( {a} \right)`: `<mi>a</mi>`,
	} {
		out, err := convertErrWithin(t, tex)
		require.NoError(t, err, tex)
		assert.Contains(t, out, want, tex)
	}
}

// Each \left( or \begin closing outside its {group} used to double the
// parse work: a few hundred bytes cost seconds and gigabytes. Rejected now,
// the cost is linear in the input.
func TestCrossingGroupCostIsLinear(t *testing.T) {
	nested := func(open, close string, k int) string {
		return strings.Repeat(open, k) + "x" + strings.Repeat(close, k)
	}
	for _, unit := range [][2]string{
		{`{\left(`, `}\right)`},
		{`{\begin{matrix}`, `}\end{matrix}`},
		{`\left({`, `\right)}`},
	} {
		best := func(tex string) time.Duration {
			d := time.Duration(1<<63 - 1)
			for range 3 {
				runtime.GC()
				start := time.Now()
				convertErrWithin(t, tex)
				d = min(d, time.Since(start))
			}
			return d
		}
		// Large enough for the timings to dominate timer noise.
		small, large := nested(unit[0], unit[1], 256), nested(unit[0], unit[1], 1024)
		_, err := convertErrWithin(t, large)
		assert.Error(t, err, unit[0])
		ratio := float64(best(large)) / float64(best(small))
		assert.Less(t, ratio, 16.0, "%q: 4x the input took %.1fx the time", unit[0], ratio)
	}
}

// Reading past the end of a buffer consumes it and fails, so a caller that
// ungets after the failure does not rewind onto the token it came from.
func TestGetNextNPastEnd(t *testing.T) {
	toks, err := tokenize([]rune("a b"))
	require.NoError(t, err)
	b := NewTokenBuffer(toks)
	b.Advance()
	got, err := b.GetNextN(5)
	assert.ErrorIs(t, err, ErrTokenBufferEnd)
	assert.Equal(t, "b", strings.TrimSpace(StringifyTokens(got.Expr)))
	assert.True(t, b.Empty())
	b.Unget()
	assert.Equal(t, 1, b.idx)

	// Skipping whitespace first cannot read past the end either.
	b = NewTokenBuffer(toks[:2])
	b.Advance()
	_, err = b.GetNextN(1, true)
	assert.ErrorIs(t, err, ErrTokenBufferEnd)
}

// A sub-buffer cannot be extended past its end into the tokens of the
// buffer it was cut from.
func TestSubBufferCapacity(t *testing.T) {
	toks, err := tokenize([]rune("{a}b[c]d"))
	require.NoError(t, err)
	b := NewTokenBuffer(toks)
	group, err := b.GetNextExpr()
	require.NoError(t, err)
	assert.Equal(t, len(group.Expr), cap(group.Expr))
	b.Advance()
	opts, err := b.GetOptions()
	require.NoError(t, err)
	assert.Equal(t, len(opts.Expr), cap(opts.Expr))
	n, err := b.GetNextN(1)
	require.NoError(t, err)
	assert.Equal(t, len(n.Expr), cap(n.Expr))
}

// A style switch applies up to the end of its cell, but the cells of an
// environment or a fence after it belong to that group, not to the switch.
func TestSwitchKeepsTheCellsOfAGroupAfterIt(t *testing.T) {
	for _, tex := range []string{
		`\displaystyle\begin{pmatrix}a&b\end{pmatrix}`,
		`\small\begin{matrix}a&b\end{matrix}`,
		`\color{red}\begin{matrix}a&b\end{matrix}`,
		`\begin{matrix}\displaystyle\left(\begin{matrix}a&b\end{matrix}\right)&c\end{matrix}`,
	} {
		out, err := convertErrWithin(t, tex)
		require.NoError(t, err, tex)
		assert.NotContains(t, out, "<mo>&amp;</mo>", tex)
		assert.Contains(t, out, `<mi>b</mi></mtd>`, tex)
	}
	// The switch still ends at the cell it is in.
	out, err := convertErrWithin(t, `\begin{matrix}\displaystyle a&b\end{matrix}`)
	require.NoError(t, err)
	assert.Contains(t, out, `<mtd><mi>b</mi></mtd>`)
}
