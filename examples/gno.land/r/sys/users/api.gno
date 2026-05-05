package users

// IsNameTaken reports whether the exact-string name exists in nameStore.
// Returns true for any name ever registered, including:
//
//   - active registrations
//   - tombstoned (deleted) users' names — Delete() sets `deleted=true`
//     but does not remove the nameStore entry (anti-revival policy)
//   - old aliases from renames — UpdateName inserts the new name
//     alongside the old; the old key stays (anti-rename-squat policy)
//
// In short: IsNameTaken(name) answers "would RegisterUser(name, _)
// fail with ErrNameTaken?" — same answer for active, deleted, or
// aliased-away names. Pairs with IsCanonicalTaken (canonical-match)
// and ResolveName (active-current-user lookup with full UserData).
//
// No canonicalization is applied. For controllers that want exact-
// match uniqueness without pulling in canonical-collision logic.
func IsNameTaken(name string) bool {
	return nameStore.Has(name)
}

// IsCanonicalTaken reports whether the given name's canonical form is
// already registered. Pass the raw name; canonicalization is applied
// internally. The first return is the original (non-canonical) name
// that owns the canonical key, for UX in collision messages.
//
// When a bypass write (RegisterUserIgnoreCanonical or the bypass path
// through ProposeRegisterUser/ProposeUpdateName) overwrites a prior
// canonical entry, this returns the most-recently-written original.
func IsCanonicalTaken(name string) (existing string, taken bool) {
	v, ok := canonicalStore.Get(Canonicalize(name))
	if !ok {
		return "", false
	}
	return v.(string), true
}
