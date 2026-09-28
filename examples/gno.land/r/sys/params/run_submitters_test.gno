package params

import (
	"chain"
	"testing"

	"gno.land/p/nt/uassert/v0"
)

const testRelay = "gno.land/r/test/relay"

// TestProposeSetRunSubmittersRefusesRelayingRealm covers a relayed request being refused.
func TestProposeSetRunSubmittersRefusesRelayingRealm(cur realm, t *testing.T) {
	resetDelegation()
	defer resetDelegation()

	testing.SetRealm(testing.NewCodeRealm(testRelay))
	uassert.AbortsContains(t, cur, "relaying realm", func() {
		ProposeSetRunSubmitters(cross(cur), []string{string(chain.PackageAddress(testRelay))})
	})
	uassert.AbortsContains(t, cur, "relaying realm", func() {
		ProposeSetRunSubmitters(cross(cur), []string{})
	})
}
