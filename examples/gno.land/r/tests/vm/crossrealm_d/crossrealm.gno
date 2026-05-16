package crossrealm_d

// Simple stateful realm for cross-realm consistency tests.
// Separated from crossrealm_b to avoid perturbing its object IDs.
//
// Contains both crossing and non-crossing setters so tests can
// demonstrate that non-crossing calls from another realm cannot
// silently mutate state via assign+recover.

var counter int

func init() {
	counter = 100
}

// SetCounter: non-crossing. Calling this cross-realm triggers
// the readonly check because it directly assigns a package var.
func SetCounter(n int) {
	counter = n
}

// SetCounterCrossing: crossing version. This works correctly
// cross-realm because the caller enters this realm's context.
func SetCounterCrossing(cur realm, n int) {
	counter = n
}

func GetCounter(cur realm) int {
	return counter
}

// DoubleCounter reads counter and doubles it. Used to show that
// if counter were silently corrupted in memory, subsequent crossing
// calls would act on the wrong value.
func DoubleCounter(cur realm) int {
	counter = counter * 2
	return counter
}

// MutateBytes mutates the first byte of bz to 0xff. Used to test
// whether a foreign realm can write to caller-allocated bytes.
// Under PLAN3 storage=authority, bz has PkgID=caller; Layer 1 borrow
// here makes m.Realm=crossrealm_d (declaring realm); the write site
// (bz[0] = ...) should fire readonly because bz.PkgID != m.Realm.ID.
func MutateBytes(bz []byte) {
	bz[0] = 0xff
}
