package grc20

import (
	mm "gno.land/p/labs/metamodel"
)

const (
	Version = "v0"

	// Fungible Objects
	Token = "$token"
	Allow = "$allow"
)

var (
	modelOpts = map[string]any{
		"transfer_amount":   10,
		"approve_amount":    400,
		"spend_amount":      100,
		"mint_amount":       5,
		"burn_amount":       2,
		"supply_initial":    0,
		"owner_balance":     1100,
		"recipient_balance": 0,
		"allowance_initial": 0,
		"minter_initial":    1,
		"objects":           []string{Token, Allow}, // "Token" oriented view
	}
)

func init() {
	modelOpts["objects"] = []string{Allow, Token} // invert view - "Allow" oriented view
}

// GRC20 returns a single model that matches GRC20 semantics.
func GRC20(opts map[string]any) *mm.Model {

	// Helper to get integer options with defaults
	getIntOpt := func(key string, def int) int64 {
		if val, ok := opts[key]; ok {
			if intval, ok := val.(int); ok {
				return int64(intval)
			}
			if intval, ok := val.(int64); ok {
				return intval
			}
		}
		return int64(def)
	}

	transferAmt := getIntOpt("transfer_amount", 10)
	approveAmt := getIntOpt("approve_amount", 400)
	spendAmt := getIntOpt("spend_amount", 100)
	mintAmt := getIntOpt("mint_amount", 5)
	burnAmt := getIntOpt("burn_amount", 2)

	om := mm.New(opts["objects"].([]string))

    //	supplyInit := T(0, 0) // $allow, $token
	supplyInit := om.T(getIntOpt("supply_initial", 0), Token)

	ownerInit := om.T(getIntOpt("owner_balance", 1100), Token)
	recipientInit := om.T(getIntOpt("recipient_balance", 0), Token)

	allowInit := om.T(getIntOpt("allowance_initial", 0), Allow)
	minterInit := om.T(getIntOpt("minter_initial", 1), Allow)
	spenderInit := om.T(1, Allow) // dummy spender to enable transferFrom

	places := map[string]mm.Place{
		"$owner":                {Initial: ownerInit, X: 43, Y: 266},
		"$recipient":            {Initial: recipientInit, X: 782, Y: 442},
		"$spender":              {Initial: spenderInit, X: 44, Y: 443},
		"allow(owner->spender)": {Initial: allowInit, X: 525, Y: 275},
		"supply":                {Initial: supplyInit, X: 513, Y: 111},
		"$burned":               {X: 764, Y: 111},
		"$minter":               {Initial: minterInit, X: 50, Y: 113},
		"$void":                 {X: 941, Y: 274},
	}

	transitions := map[string]mm.Transition{
		"transfer":       {X: 349, Y: 444},
		"approveAdd":     {X: 345, Y: 277},
		"approveZero":    {X: 680, Y: 275},
		"transferFrom":   {X: 619, Y: 444},
		"spendAllowance": {X: 799, Y: 275},
		"mint":           {X: 347, Y: 185},
		"burn":           {X: 345, Y: 113},
	}

	arc := func(source, target string, weight int64, object string) mm.Arrow {
		a := mm.Arc{Source: source, Target: target, Weight: weight, Object: object}
		return om.Arrow(a)
	}

	inhibit := func(source, target string, weight int64, object string) mm.Arrow {
		a := mm.Arc{Source: source, Target: target, Weight: weight, Inhibit: true, Object: object}
		return om.Arrow(a)
	}

	arrows := []mm.Arrow{
		arc("$owner", "transfer", transferAmt, Token),
		arc("transfer", "$recipient", transferAmt, Token),
		inhibit("approveAdd", "$owner", 0, Allow),

		arc("approveAdd", "allow(owner->spender)", approveAmt, Allow),
		arc("allow(owner->spender)", "approveZero", spendAmt, Allow),
		arc("approveZero", "$void", spendAmt, Allow),
		arc("$owner", "transferFrom", transferAmt, Token),
		arc("allow(owner->spender)", "transferFrom", transferAmt, Allow),
		arc("transferFrom", "$recipient", transferAmt, Token),
		arc("allow(owner->spender)", "spendAllowance", spendAmt, Allow),
		arc("spendAllowance", "$void", spendAmt, Allow),
		inhibit("mint", "$minter", 0, Allow),

		arc("mint", "$owner", mintAmt, Token),
		arc("mint", "supply", mintAmt, Token),
		arc("$owner", "burn", burnAmt, Token),
		arc("supply", "burn", burnAmt, Token),
		arc("burn", "$burned", burnAmt, Token),
		inhibit("transferFrom", "$spender", 0, Allow),
	}

	return mm.New(
		om.Objects,
		places,
		transitions,
		arrows,
		map[string]any{"version": Version, "name": "GRC20"},
	)
}

/*
Categorical view of `allow(owner->spender)`

- Implementation note:
    We model Allow(o,s) as its own place named "allow(owner->spender)" — a
    discrete-index view of the profunctor. A colored-token variant would store
    pairs (o,s) as token color in a single "$allow" place; the current layout is
    a clear, didactic approximation.

- Profunctor (enriched in the commutative monoid (N,+,0)):
    Allow : Addr^op × Addr -> (N,+,0)
    Each pair (owner, spender) maps to a natural-number allowance.
    The map is a sparse matrix representation of this profunctor.

- Pointwise (natural) operations:
    ApproveAdd(o,s,a):   Allow(o,s) += a
    ApproveZero(o,s):    Allow(o,s)  = 0
    SpendAllowance(o,s,a): Allow(o,s) = max(Allow(o,s) - a, 0)

- Atomic TransferFrom:
    Requires Token(o) and Allow(o,s) as simultaneous inputs, and produces Token(r).
    Effect: Token moves owner->recipient by 'a' while Allow(o,s) decreases by 'a'
    in the same transition. This is the categorical “commuting square” that
    enforces atomicity.

- Yoneda-ish slices:
    Fix owner o:   Allow(o, -) : Addr -> (N,+,0)   (copresheaf: budgets by spender)
    Fix spender s: Allow(-, s) : Addr^op -> (N,+,0) (presheaf: who grants me how much)

*/
