package antispam

import (
	"chain/runtime"

	"gno.land/p/nt/avl"

	engine "gno.land/p/gnoland/antispam"
)

// ---- Per-address reputation state

// addrReputation holds moderation data tracked by the realm.
// Chain data (AccountAgeDays, Balance, HasUsername) must still
// be provided by the caller since GnoVM cannot query other
// addresses' on-chain state.
type addrReputation struct {
	totalAccepted int
	flaggedCount  int
	banCount      int
}

var (
	reputations    = avl.NewTree() // address string -> *addrReputation
	trustedCallers = avl.NewTree() // address string -> true (registered caller realms)
)

// getOrCreate returns the reputation entry for addr, creating one if needed.
func getOrCreate(addr address) *addrReputation {
	key := addr.String()
	raw, ok := reputations.Get(key)
	if ok {
		return raw.(*addrReputation)
	}
	rep := &addrReputation{}
	reputations.Set(key, rep)
	return rep
}

// assertTrustedCaller panics if the caller is not a registered realm
// or the admin calling directly. Prevents untrusted realms from
// poisoning reputation data.
func assertTrustedCaller() {
	prev := runtime.PreviousRealm()
	// Direct admin calls are always allowed
	if prev.IsUser() && prev.Address() == adminAddr {
		return
	}
	// Realm callers must be registered
	if _, ok := trustedCallers.Get(prev.Address().String()); !ok {
		panic(errUnauthorized)
	}
}

// ---- Public: reputation recording (trusted callers only)

// RecordAccepted increments the accepted-content count for an address.
// Caller realms should call this when content is accepted into the system.
func RecordAccepted(_ realm, addr address) {
	assertTrustedCaller()
	rep := getOrCreate(addr)
	rep.totalAccepted++
}

// RecordFlag increments the flag count for an address.
// Caller realms should call this when content is flagged by moderators.
func RecordFlag(_ realm, addr address) {
	assertTrustedCaller()
	rep := getOrCreate(addr)
	rep.flaggedCount++
}

// RecordBan increments the ban count for an address.
// Ban history persists even after unbanning.
func RecordBan(_ realm, addr address) {
	assertTrustedCaller()
	rep := getOrCreate(addr)
	rep.banCount++
}

// ---- Public: reputation queries

// GetReputation returns the moderation counters for an address.
// Chain data fields (AccountAgeDays, Balance, HasUsername) are left
// at zero - callers must fill those from their own context.
// Returns zero values for unknown addresses.
func GetReputation(addr address) engine.ReputationData {
	key := addr.String()
	raw, ok := reputations.Get(key)
	if !ok {
		return engine.ReputationData{}
	}
	rep := raw.(*addrReputation)
	return engine.ReputationData{
		TotalAccepted: rep.totalAccepted,
		FlaggedCount:  rep.flaggedCount,
		BanCount:      rep.banCount,
	}
}

// ReputationCount returns the number of tracked addresses.
func ReputationCount() int {
	return reputations.Size()
}

// TrustedCallerCount returns the number of registered caller realms.
func TrustedCallerCount() int {
	return trustedCallers.Size()
}

// ---- Admin: reputation management

// AdminResetReputation clears all moderation data for an address.
func AdminResetReputation(_ realm, addr address) {
	assertAdmin()
	reputations.Remove(addr.String())
}

// ---- Admin: trusted caller management

// AdminRegisterCaller adds a realm address to the trusted caller list.
// Only registered callers can record reputation actions.
func AdminRegisterCaller(_ realm, callerAddr address) {
	assertAdmin()
	trustedCallers.Set(callerAddr.String(), true)
}

// AdminRemoveCaller removes a realm address from the trusted caller list.
func AdminRemoveCaller(_ realm, callerAddr address) {
	assertAdmin()
	trustedCallers.Remove(callerAddr.String())
}
