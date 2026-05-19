package boards

// PermissionSet defines a type to store any number of permissions.
type PermissionSet []uint64

// NewPermissionSet creates a new PermissionSet containing the given permissions.
func NewPermissionSet(perms ...Permission) PermissionSet {
	if len(perms) == 0 {
		return nil
	}

	// Find max permission value to calculate slice size.
	// This allows any number of permissions to be assigned in any order.
	var max Permission
	for _, p := range perms {
		if p > max {
			max = p
		}
	}

	s := make(PermissionSet, int(max)/64+1)
	for _, p := range perms {
		// Calculate the index within the set where the permission should be defined.
		// Each item in the set can contain 64 permissions, for example:
		// - Item 0: permissions 0 to 63
		// - Item 1: permissions 64 to 127
		idx := int(p) / 64

		// Turn on the bit that matches the permission, ranging from bit 0 to 63
		s[idx] |= 1 << (uint(p) % 64)
	}
	return s
}

// Has checks if a permission is in the set.
func (s PermissionSet) Has(p Permission) bool {
	idx := int(p) / 64
	if idx >= len(s) {
		return false
	}

	// Check if the bit for the current permission is on
	return s[idx]&(1<<(uint(p)%64)) != 0
}

// IsEmpty reports whether the set contains no permissions.
func (s PermissionSet) IsEmpty() bool {
	for _, v := range s {
		if v != 0 {
			return false
		}
	}
	return true
}
