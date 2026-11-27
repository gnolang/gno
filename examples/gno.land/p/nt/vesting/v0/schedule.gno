package vesting

import "strings"

// Validate checks a schedule. A linear curve has a zero period. A step
// curve's period is positive and divides Duration, so every step is the
// same length and the last step lands on the end.
func (t Terms) Validate() error {
	if t.Total <= 0 {
		return ErrTotal
	}
	if t.Start < 0 {
		return ErrStart
	}
	if t.Duration <= 0 {
		return ErrDuration
	}
	if t.Cliff < 0 || t.Cliff > t.Duration {
		return ErrCliff
	}
	if _, err := addInt(t.Start, t.Duration); err != nil {
		return ErrMath
	}
	switch t.Curve {
	case CurveLinear:
		if t.Period != 0 {
			return ErrPeriod
		}
	case CurveStep:
		if t.Period <= 0 || t.Period > t.Duration || t.Duration%t.Period != 0 {
			return ErrPeriod
		}
	default:
		return ErrCurve
	}
	return nil
}

// End is the unix second at which the grant is fully vested.
func (t Terms) End() (int64, error) {
	if err := t.Validate(); err != nil {
		return 0, err
	}
	return addInt(t.Start, t.Duration)
}

// VestedAmount is the amount unlocked at now.
//
// Nothing vests before Start, or before Start+Cliff. At and after
// Start+Duration the whole Total is vested, so rounding dust is not
// left behind. In between, the amount is floor(Total * elapsed / span),
// which favors the vault. On a step curve the span is the number of
// periods and elapsed is the number of periods completed since Start.
// A cliff that falls inside a period releases every period already
// completed and none of the period still open.
func VestedAmount(t Terms, now int64) (int64, error) {
	if err := t.Validate(); err != nil {
		return 0, err
	}
	if now < 0 {
		return 0, ErrTime
	}
	if now < t.Start {
		return 0, nil
	}
	end, err := addInt(t.Start, t.Duration)
	if err != nil {
		return 0, err
	}
	if t.Cliff > 0 {
		cliffAt, cerr := addInt(t.Start, t.Cliff)
		if cerr != nil {
			return 0, cerr
		}
		if now < cliffAt {
			return 0, nil
		}
	}
	if now >= end {
		return t.Total, nil
	}
	elapsed, err := subInt(now, t.Start)
	if err != nil {
		return 0, err
	}
	switch t.Curve {
	case CurveLinear:
		return mulDivFloor(t.Total, elapsed, t.Duration)
	case CurveStep:
		steps := t.Duration / t.Period
		done := elapsed / t.Period
		if done <= 0 {
			return 0, nil
		}
		if done >= steps {
			return t.Total, nil
		}
		return mulDivFloor(t.Total, done, steps)
	default:
		return 0, ErrCurve
	}
}

func validLabel(s string) bool {
	if len(s) > MaxLabelLen {
		return false
	}
	for _, r := range s {
		if r < 0x20 || r == 0x7f {
			return false
		}
	}
	return true
}

func validDenom(denom string) bool {
	if len(denom) < 3 || len(denom) > MaxDenomLen {
		return false
	}
	c0 := denom[0]
	if (c0 < 'a' || c0 > 'z') && c0 != '/' {
		return false
	}
	for i := 1; i < len(denom); i++ {
		switch c := denom[i]; {
		case c >= 'a' && c <= 'z', c >= '0' && c <= '9':
		case c == '_', c == '.', c == ':', c == '/', c == '-':
		default:
			return false
		}
	}
	return true
}

func validToken(key string) bool {
	if key == "" || len(key) > MaxTokenLen {
		return false
	}
	if strings.Contains(key, "#") {
		return false
	}
	for _, r := range key {
		if r < 0x20 || r == 0x7f || r == ' ' {
			return false
		}
	}
	return true
}

func (a Asset) validate() error {
	switch a.Kind {
	case AssetCoin:
		if a.Token != "" || !validDenom(a.Denom) {
			return ErrDenom
		}
	case AssetGRC20:
		if a.Denom != "" || !validToken(a.Token) {
			return ErrToken
		}
	default:
		return ErrAsset
	}
	return nil
}
