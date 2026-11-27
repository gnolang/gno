package vesting

// Limits keep one grant's stored strings inside the bank's denom
// rules and a page that does not walk attacker-sized text.
const (
	MaxLabelLen = 64
	// MaxDenomLen is std.MaxDenomLength: "/" + 256-byte pkg path + ":" + 16.
	MaxDenomLen = 274
	MaxTokenLen = 320

	CurveLinear Curve = 1
	CurveStep   Curve = 2

	AssetCoin  AssetKind = 1
	AssetGRC20 AssetKind = 2
)

// Curve selects how tokens unlock after the cliff.
type Curve int

// String returns a stable curve name.
func (c Curve) String() string {
	switch c {
	case CurveLinear:
		return "linear"
	case CurveStep:
		return "step"
	default:
		return "unknown"
	}
}

// ParseCurve accepts the integers MsgCall can pass.
func ParseCurve(n int64) (Curve, error) {
	switch Curve(n) {
	case CurveLinear, CurveStep:
		return Curve(n), nil
	default:
		return 0, ErrCurve
	}
}

// AssetKind selects a native coin denom or a GRC20 registry key.
type AssetKind int

// String returns a stable asset name.
func (k AssetKind) String() string {
	switch k {
	case AssetCoin:
		return "coin"
	case AssetGRC20:
		return "grc20"
	default:
		return "unknown"
	}
}

// Asset is one denom or one registry key. A grant locks a single asset.
type Asset struct {
	Kind  AssetKind
	Denom string
	Token string
}

// Terms is the schedule. Start is a unix second. Cliff, Duration, and
// Period are durations in seconds, measured from Start. Period is zero
// on a linear curve. Total is the whole grant, in the asset's base units.
type Terms struct {
	Start    int64
	Cliff    int64
	Duration int64
	Period   int64
	Total    int64
	Curve    Curve
}

// GrantView is a copy of a grant plus the amounts at Now.
type GrantView struct {
	ID          int64
	Admin       address
	Beneficiary address
	Label       string
	Asset       Asset
	Terms       Terms
	Claimed     int64
	Returned    int64
	Revocable   bool
	Revoked     bool
	RevokedAt   int64
	Address     address
	Vested      int64
	Releasable  int64
}
