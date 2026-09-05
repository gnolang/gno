package grc721reg

// Value mirrors grc20reg's register event for cross-registry consistency.
const RegisterEvent = "register"

const UnregisterEvent = "unregister"

const maxSlugLen = 128

// The index is open to any realm, so what one page renders is bounded here
// rather than by whoever registers most.
const (
	pageSize   = 20
	maxBadges  = 8
	maxKindLen = 64
)

// Kinds also reach the register event, where chain.Emit's own cap would
// otherwise be the only ceiling.
const maxEventKinds = 16
