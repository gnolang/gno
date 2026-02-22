package antispam

import (
	"gno.land/p/nt/avl"

	engine "gno.land/p/gnoland/antispam"
)

// ---- Per-address reputation state

// addrReputation holds moderation data tracked by the realm.
// Chain data (AccountAgeDays, Balance, HasUsername) must still
// be provided by the caller since GnoVM cannot query other
// addresses' on-chain state.
type addrReputation struct {
	totalPosts   int
	flaggedCount int
	banCount     int
}

var reputations = avl.NewTree() // address string -> *addrReputation

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

// ---- Public: reputation recording

// RecordPost increments the post count for an address.
// Caller realms should call this when a post is accepted.
func RecordPost(_ realm, addr address) {
	rep := getOrCreate(addr)
	rep.totalPosts++
}

// RecordFlag increments the flag count for an address.
// Caller realms should call this when a post is flagged by moderators.
func RecordFlag(_ realm, addr address) {
	rep := getOrCreate(addr)
	rep.flaggedCount++
}

// RecordBan increments the ban count for an address.
// Ban history persists even after unbanning.
func RecordBan(_ realm, addr address) {
	rep := getOrCreate(addr)
	rep.banCount++
}

// RecordUnban is a no-op for ban count (history persists) but
// can be extended for active-ban tracking in the future.
func RecordUnban(_ realm, addr address) {
	// Ban count is historical and not decremented.
	// Ensure the address exists in the tree for consistency.
	getOrCreate(addr)
}

// ---- Public: reputation queries

// GetReputation returns the moderation data for an address.
// Returns zero values for unknown addresses.
func GetReputation(addr address) engine.ReputationData {
	key := addr.String()
	raw, ok := reputations.Get(key)
	if !ok {
		return engine.ReputationData{}
	}
	rep := raw.(*addrReputation)
	return engine.ReputationData{
		TotalPosts:   rep.totalPosts,
		FlaggedCount: rep.flaggedCount,
		BanCount:     rep.banCount,
	}
}

// ReputationCount returns the number of tracked addresses.
func ReputationCount() int {
	return reputations.Size()
}

// ---- Admin: reputation management

// AdminResetReputation clears all moderation data for an address.
func AdminResetReputation(_ realm, addr address) {
	assertAdmin()
	reputations.Remove(addr.String())
}
