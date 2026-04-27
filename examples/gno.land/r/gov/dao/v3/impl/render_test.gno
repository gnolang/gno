package impl

import (
	"testing"

	"gno.land/p/nt/urequire/v0"
)

// TestRenderMembershipBadge verifies that the non-member badge is displayed correctly
func TestRenderMembershipBadge(t *testing.T) {
	loadMembers()

	// Test member (should not show badge)
	badge := renderMembershipBadge(m1)
	urequire.Equal(t, "", badge, "Members should not have a badge")

	// Test non-member (should show badge)
	badge = renderMembershipBadge(noMember)
	urequire.Equal(t, "`[non-member]`", badge, "Non-members should have a badge")
}
