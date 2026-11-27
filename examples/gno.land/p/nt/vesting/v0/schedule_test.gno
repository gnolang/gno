package vesting

import (
	"math"
	"testing"

	"gno.land/p/nt/uassert/v0"
)

func TestVestedAmountBoundaries(t *testing.T) {
	const (
		start = int64(1_000_000)
		total = int64(1000)
	)
	linear := Terms{Start: start, Cliff: 100, Duration: 1000, Total: total, Curve: CurveLinear}
	step := Terms{Start: start, Cliff: 250, Duration: 1000, Period: 100, Total: total, Curve: CurveStep}
	cases := []struct {
		name string
		term Terms
		now  int64
		want int64
	}{
		{"before start", linear, start - 1, 0},
		{"exactly start, still in cliff", linear, start, 0},
		{"one second before cliff", linear, start + 99, 0},
		{"exactly cliff", linear, start + 100, 100},
		{"midway", linear, start + 500, 500},
		{"one before end", linear, start + 999, 999},
		{"exactly end", linear, start + 1000, 1000},
		{"long after end", linear, start + 50_000, 1000},
		{"cliff zero at start", Terms{Start: start, Duration: 1000, Total: total, Curve: CurveLinear}, start, 0},
		{"cliff zero one second", Terms{Start: start, Duration: 1000, Total: total, Curve: CurveLinear}, start + 1, 1},
		{"cliff equals duration before end", Terms{Start: start, Cliff: 1000, Duration: 1000, Total: total, Curve: CurveLinear}, start + 999, 0},
		{"cliff equals duration at end", Terms{Start: start, Cliff: 1000, Duration: 1000, Total: total, Curve: CurveLinear}, start + 1000, 1000},
		{"duration one before", Terms{Start: start, Duration: 1, Total: total, Curve: CurveLinear}, start, 0},
		{"duration one at end", Terms{Start: start, Duration: 1, Total: total, Curve: CurveLinear}, start + 1, 1000},
		{"tiny total stays dust until end", Terms{Start: start, Duration: 1000, Total: 1, Curve: CurveLinear}, start + 999, 0},
		{"tiny total released at end", Terms{Start: start, Duration: 1000, Total: 1, Curve: CurveLinear}, start + 1000, 1},
		{"step before cliff", step, start + 249, 0},
		{"step at cliff releases completed periods", step, start + 250, 200},
		{"step just before a boundary", step, start + 299, 200},
		{"step on a boundary", step, start + 300, 300},
		{"step one before end", step, start + 999, 900},
		{"step at end", step, start + 1000, 1000},
		{"step cliff zero at start", Terms{Start: start, Duration: 1000, Period: 100, Total: total, Curve: CurveStep}, start, 0},
		{"step duration one", Terms{Start: start, Duration: 1, Period: 1, Total: 7, Curve: CurveStep}, start + 1, 7},
		{"half total rounds down", Terms{Start: start, Duration: 3, Total: 10, Curve: CurveLinear}, start + 1, 3},
		{"half total rounds down again", Terms{Start: start, Duration: 3, Total: 10, Curve: CurveLinear}, start + 2, 6},
	}
	for _, tc := range cases {
		got, err := VestedAmount(tc.term, tc.now)
		uassert.NoError(t, err, tc.name)
		uassert.Equal(t, tc.want, got, tc.name)
	}
}

func TestVestedAmountHuge(t *testing.T) {
	const start = int64(10)
	total := int64(math.MaxInt64)
	term := Terms{Start: start, Duration: 100, Total: total, Curve: CurveLinear}
	cases := []struct {
		elapsed int64
		want    int64
	}{
		{0, 0},
		{1, 92233720368547758},
		{50, 4611686018427387903},
		{99, 9131138316486228048},
		{100, total},
		{1000, total},
	}
	for _, tc := range cases {
		got, err := VestedAmount(term, start+tc.elapsed)
		uassert.NoError(t, err)
		uassert.Equal(t, tc.want, got)
	}
	third := Terms{Start: start, Duration: 3, Total: total, Curve: CurveLinear}
	got, err := VestedAmount(third, start+1)
	uassert.NoError(t, err)
	uassert.Equal(t, int64(3074457345618258602), got)
	got, err = VestedAmount(third, start+2)
	uassert.NoError(t, err)
	uassert.Equal(t, int64(6148914691236517204), got)

	step := Terms{Start: start, Cliff: 0, Duration: 4, Period: 2, Total: total, Curve: CurveStep}
	got, err = VestedAmount(step, start+2)
	uassert.NoError(t, err)
	uassert.Equal(t, int64(4611686018427387903), got)
	got, err = VestedAmount(step, start+4)
	uassert.NoError(t, err)
	uassert.Equal(t, total, got)
}

func TestTermsReject(t *testing.T) {
	ok := Terms{Start: 10, Cliff: 1, Duration: 4, Total: 8, Curve: CurveLinear}
	cases := []struct {
		name string
		term Terms
		want error
	}{
		{"zero total", Terms{Start: 1, Duration: 1, Curve: CurveLinear}, ErrTotal},
		{"negative total", Terms{Start: 1, Duration: 1, Total: -1, Curve: CurveLinear}, ErrTotal},
		{"negative start", Terms{Start: -1, Duration: 1, Total: 1, Curve: CurveLinear}, ErrStart},
		{"zero duration", Terms{Start: 0, Total: 1, Curve: CurveLinear}, ErrDuration},
		{"negative duration", Terms{Start: 0, Duration: -5, Total: 1, Curve: CurveLinear}, ErrDuration},
		{"negative cliff", Terms{Start: 0, Cliff: -1, Duration: 4, Total: 1, Curve: CurveLinear}, ErrCliff},
		{"cliff past duration", Terms{Start: 0, Cliff: 5, Duration: 4, Total: 1, Curve: CurveLinear}, ErrCliff},
		{"start plus duration overflows", Terms{Start: math.MaxInt64, Duration: 1, Total: 1, Curve: CurveLinear}, ErrMath},
		{"linear with a period", Terms{Start: 0, Duration: 4, Period: 2, Total: 1, Curve: CurveLinear}, ErrPeriod},
		{"step period zero", Terms{Start: 0, Duration: 4, Total: 1, Curve: CurveStep}, ErrPeriod},
		{"step period does not divide", Terms{Start: 0, Duration: 4, Period: 3, Total: 1, Curve: CurveStep}, ErrPeriod},
		{"step period longer than duration", Terms{Start: 0, Duration: 4, Period: 5, Total: 1, Curve: CurveStep}, ErrPeriod},
		{"unknown curve", Terms{Start: 0, Duration: 4, Total: 1, Curve: 9}, ErrCurve},
		{"negative now", ok, ErrTime},
	}
	for _, tc := range cases {
		if tc.want == ErrTime {
			_, err := VestedAmount(tc.term, -1)
			uassert.ErrorIs(t, err, ErrTime, tc.name)
			continue
		}
		err := tc.term.Validate()
		uassert.ErrorIs(t, err, tc.want, tc.name)
		_, err = VestedAmount(tc.term, 0)
		uassert.ErrorIs(t, err, tc.want, tc.name)
	}
	end, err := ok.End()
	uassert.NoError(t, err)
	uassert.Equal(t, int64(14), end)
	_, err = Terms{}.End()
	uassert.ErrorIs(t, err, ErrTotal)
}

func TestParseCurve(t *testing.T) {
	c, err := ParseCurve(int64(CurveLinear))
	uassert.NoError(t, err)
	uassert.Equal(t, "linear", c.String())
	c, err = ParseCurve(int64(CurveStep))
	uassert.NoError(t, err)
	uassert.Equal(t, "step", c.String())
	_, err = ParseCurve(0)
	uassert.ErrorIs(t, err, ErrCurve)
	uassert.Equal(t, "unknown", Curve(0).String())
	uassert.Equal(t, "coin", AssetCoin.String())
	uassert.Equal(t, "grc20", AssetGRC20.String())
	uassert.Equal(t, "unknown", AssetKind(0).String())
}

func TestMulDivFloor(t *testing.T) {
	got, err := mulDivFloor(10, 1, 3)
	uassert.NoError(t, err)
	uassert.Equal(t, int64(3), got)
	got, err = mulDivFloor(0, 5, 1)
	uassert.NoError(t, err)
	uassert.Equal(t, int64(0), got)
	got, err = mulDivFloor(5, 0, 1)
	uassert.NoError(t, err)
	uassert.Equal(t, int64(0), got)
	_, err = mulDivFloor(-1, 1, 1)
	uassert.ErrorIs(t, err, ErrMath)
	_, err = mulDivFloor(1, -1, 1)
	uassert.ErrorIs(t, err, ErrMath)
	_, err = mulDivFloor(1, 1, 0)
	uassert.ErrorIs(t, err, ErrMath)
	_, err = mulDivFloor(math.MaxInt64, math.MaxInt64, 1)
	uassert.ErrorIs(t, err, ErrMath)
	_, err = mulDivFloor(math.MaxInt64, math.MaxInt64, 1<<62)
	uassert.ErrorIs(t, err, ErrMath)
	_, err = addInt(math.MaxInt64, 1)
	uassert.ErrorIs(t, err, ErrMath)
	_, err = subInt(0, 1)
	uassert.ErrorIs(t, err, ErrMath)
}
