package grc20

import (
    mm "gno.land/p/labs/metamodel"
)

const (
    Version = "v0"

    // Objects
    Token = "$token"
    Allow = "$allow"
)

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

    // Objects
    objects := []string{Token, Allow}

    objectModel := mm.New(objects)

    // Tunables
    transferAmt := int64(getIntOpt("transfer_amount", 10))
    approveAmt := int64(getIntOpt("approve_amount", 400))
    spendAmt := int64(getIntOpt("spend_amount", 100))
    mintAmt := int64(getIntOpt("mint_amount", 5))
    burnAmt := int64(getIntOpt("burn_amount", 2))

    // Initial values
    supplyInit := objectModel.T(getIntOpt("supply_initial", 0), Token)
    ownerInit := objectModel.T(getIntOpt("owner_balance", 1100), Token)
    recipientInit := objectModel.T(getIntOpt("recipient_balance", 0), Token)

    allowInit := objectModel.T(getIntOpt("allowance_initial", 0), Allow)
    minterInit := objectModel.T(getIntOpt("minter_initial", 1), Allow)

places := map[string]mm.Place{
    "$owner":               {Initial: ownerInit, X: 43, Y: 266},
    "$recipient":           {Initial: recipientInit, X: 782, Y: 442},
    "$spender":             {Initial: mm.T(1), X: 44, Y: 443},
    "$allow(owner->spender)": {Initial: allowInit, X: 525, Y: 275},
    "$supply":              {Initial: supplyInit, X: 513, Y: 111},
    "$burned":              {Initial: mm.T(0), X: 764, Y: 111},
    "$minter":              {Initial: minterInit, X: 50, Y: 113},
    "$void":                {Initial: mm.T(0), X: 941, Y: 274},
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

arrows := []mm.Arrow{
    objectModel.Arrow(mm.Arc{Source: "$owner", Target: "transfer", Weight: transferAmt, Object: Token}),
    objectModel.Arrow(mm.Arc{Source: "transfer", Target: "$recipient", Weight: transferAmt, Object: Token}),
    objectModel.Arrow(mm.Arc{Source: "approveAdd", Target: "$owner", Weight: approveAmt, Inhibit: true, Object: Allow}),
    objectModel.Arrow(mm.Arc{Source: "approveAdd", Target: "$allow(owner->spender)", Weight: approveAmt, Object: Allow}),
    objectModel.Arrow(mm.Arc{Source: "$allow(owner->spender)", Target: "approveZero", Weight: spendAmt, Object: Allow}),
    objectModel.Arrow(mm.Arc{Source: "approveZero", Target: "$void", Weight: spendAmt, Object: Allow}),
    objectModel.Arrow(mm.Arc{Source: "$owner", Target: "transferFrom", Weight: transferAmt, Object: Token}),
    objectModel.Arrow(mm.Arc{Source: "$allow(owner->spender)", Target: "transferFrom", Weight: transferAmt, Object: Allow}),
    objectModel.Arrow(mm.Arc{Source: "transferFrom", Target: "$recipient", Weight: transferAmt, Object: Token}),
    objectModel.Arrow(mm.Arc{Source: "$allow(owner->spender)", Target: "spendAllowance", Weight: spendAmt, Object: Allow}),
    objectModel.Arrow(mm.Arc{Source: "spendAllowance", Target: "$void", Weight: spendAmt, Object: Allow}),
    objectModel.Arrow(mm.Arc{Source: "mint", Target: "$minter", Inhibit: true, Object: Allow}),
    objectModel.Arrow(mm.Arc{Source: "mint", Target: "$owner", Weight: mintAmt, Object: Token}),
    objectModel.Arrow(mm.Arc{Source: "mint", Target: "$supply", Weight: mintAmt, Object: Token}),
    objectModel.Arrow(mm.Arc{Source: "$owner", Target: "burn", Weight: burnAmt, Object: Token}),
    objectModel.Arrow(mm.Arc{Source: "$supply", Target: "burn", Weight: burnAmt, Object: Token}),
    objectModel.Arrow(mm.Arc{Source: "burn", Target: "$burned", Weight: burnAmt, Object: Token}),
    objectModel.Arrow(mm.Arc{Source: "transferFrom", Target: "$spender", Inhibit: true, Object: Allow}),
}

    // Return the model
    return mm.New(
        objects,
        places,
        transitions,
        arrows,
        map[string]any{"version": Version, "name": "GRC20"},
    )
}

/*
Categorical view of `$allow(owner->spender)`

- Implementation note:
    We model Allow(o,s) as its own place named "$allow(owner->spender)" — a
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