package impl

import (
	"testing"

	"gno.land/p/nt/urequire"
)

// TestRenderMembershipBadge verifies that the non-member badge is displayed correctly
func TestRenderMembershipBadge(t *testing.T) {
	loadMembers()
	dao := NewGovDAO()

	// Test member (should not show badge)
	badge := renderMembershipBadge(m1, dao)
	urequire.Equal(t, "", badge, "Members should not have a badge")

	// Test non-member (should show badge)
	badge = renderMembershipBadge(noMember, dao)
	urequire.Equal(t, " `[non-member]`", badge, "Non-members should have a badge")
}
