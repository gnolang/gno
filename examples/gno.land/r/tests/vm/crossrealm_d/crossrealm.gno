package crossrealm_d

// Governance + admin state for cross-realm assign+recover tests.
// Separated from crossrealm_b to avoid perturbing its object IDs.

import "chain/runtime"

// Governance state
var (
	yesVotes  int
	noVotes   int
	finalized bool
)

func init() {
	yesVotes = 10
	noVotes = 90
	finalized = true
}

// SetVotes: non-crossing, should panic if called cross-realm.
func SetVotes(yes, no int) {
	yesVotes = yes
	noVotes = no
}

func GetResult(cur realm) string {
	if !finalized {
		return "pending"
	}
	if yesVotes > noVotes {
		return "approved"
	}
	return "rejected"
}

// Admin state
var admin address

func SetAdmin(newAdmin address) {
	admin = newAdmin
}

func GetAdmin(cur realm) address {
	return admin
}

func Withdraw(cur realm) {
	caller := runtime.CurrentRealm().Address()
	println("caller: ", caller)
	if caller != admin {
		panic("unauthorized")
	}
	println("===transfer...")
}
