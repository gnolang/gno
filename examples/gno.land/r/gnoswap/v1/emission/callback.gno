package emission

import (
	"gno.land/r/gnoswap/v1/gns"
)

var callbackStakerEmissionChange func(amount uint64)

func SetCallbackStakerEmissionChange(callback func(amount uint64)) {
	if callbackStakerEmissionChange != nil {
		panic("callbackStakerEmissionChange already set")
	}
	callbackStakerEmissionChange = callback
}

// Called when per-block emission is changed from the gns side.
// It does not process non-immediate emission changes, such as halving.
func callbackEmissionChange(amount uint64) {
	calculatedAmount := calculateAmount(amount, GetDistributionBpsPct(LIQUIDITY_STAKER))
	callbackStakerEmissionChange(calculatedAmount)
}

func RegisterGnsChangeCallback() {
	assertStakerOnly()
	gns.SetCallbackEmissionChange(callbackEmissionChange)
}
