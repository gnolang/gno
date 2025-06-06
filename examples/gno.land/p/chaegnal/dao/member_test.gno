package dao

import (
	"testing"

	"gno.land/p/demo/uassert"
)

// Test PendingApproval
func Test_PendingApproval_Success(t *testing.T) {
	dao := NewDAO("admin")

	uassert.NotPanics(t, func() { dao.PendingApproval("user") })
}

func Test_PendingApproval_Failure_AlreadyAwaiting(t *testing.T) {
	dao := NewDAO("admin")

	dao.PendingApproval("user")
	uassert.PanicsWithMessage(t, "already awaiting approval", func() { dao.PendingApproval("user") })
}

func Test_PendingApproval_Failure_AlreadyIn(t *testing.T) {
	dao := NewDAO("admin")

	uassert.PanicsWithMessage(t, "already whitelisted", func() { dao.PendingApproval("admin") })
}

// Test WhitelistUser
func Test_WhitelistUser_Success(t *testing.T) {
	dao := NewDAO("admin")

	dao.PendingApproval("user")
	uassert.NotPanics(t, func() { dao.WhitelistUser("admin", "user") })

	member, ok := dao.IsMember("user")
	if !ok || member.Rank != memberRank {
		t.Errorf("Expected user to be whitelisted with member rank.")
	}
}

func Test_WhitelistUser_Failure_NotAdmin(t *testing.T) {
	dao := NewDAO("admin")

	dao.PendingApproval("user")
	uassert.PanicsWithMessage(t, "only admins can whitelist members", func() { dao.WhitelistUser("user", "user") })
}

func Test_WhitelistUser_Failure_NotAwaiting(t *testing.T) {
	dao := NewDAO("admin")

	uassert.PanicsWithMessage(t, "this address is not awaiting approval", func() { dao.WhitelistUser("admin", "user") })
}

func Test_WhitelistUser_Failure_AlreadyIn(t *testing.T) {
	dao := NewDAO("admin")

	dao.PendingApproval("user")
	dao.WhitelistUser("admin", "user")

	uassert.PanicsWithMessage(t, "already whitelisted", func() { dao.PendingApproval("user") })
}

// Test RefuseAwaitingMember
func Test_RefuseAwaitingMember_Success(t *testing.T) {
	dao := NewDAO("admin")

	dao.PendingApproval("user")
	uassert.NotPanics(t, func() { dao.RefuseAwaitingMember("admin", "user") })
}

func Test_RefuseAwaitingMember_Failure(t *testing.T) {
	dao := NewDAO("admin")

	dao.PendingApproval("user")
	uassert.PanicsWithMessage(t, "only admins can refuse members", func() { dao.RefuseAwaitingMember("user", "user") })
}

// Test SetAdmin
func Test_SetAdmin_Success(t *testing.T) {
	dao := NewDAO("admin")

	dao.PendingApproval("user")
	dao.WhitelistUser("admin", "user")

	uassert.NotPanics(t, func() { dao.SetAdmin("admin", "user") })
}

func Test_SetAdmin_Failure_NotAdmin(t *testing.T) {
	dao := NewDAO("admin")

	dao.PendingApproval("user")
	dao.WhitelistUser("admin", "user")

	uassert.PanicsWithMessage(t, "only admins can promote members", func() { dao.SetAdmin("user", "user") })
}

// Test defaultAdmin
func Test_SetAdmin_Failure_NotMember(t *testing.T) {
	dao := NewDAO("admin")

	uassert.PanicsWithMessage(t, "not a member", func() { dao.defaultAdmin("user") })
}

func Test_SetAdmin_Failure_AlreadyAdmin(t *testing.T) {
	dao := NewDAO("admin")

	dao.PendingApproval("user")
	dao.WhitelistUser("admin", "user")

	dao.defaultAdmin("user")
	uassert.PanicsWithMessage(t, "member is already an admin", func() { dao.defaultAdmin("user") })
}

// Test RemoveAdminRank
func Test_RemoveAdminRank_Success(t *testing.T) {
	dao := NewDAO("admin")

	dao.PendingApproval("user")
	dao.WhitelistUser("admin", "user")
	dao.SetAdmin("admin", "user")

	uassert.NotPanics(t, func() { dao.RemoveAdminRank("admin", "user") })
}

func Test_RemoveAdminRank_Failure_CallerNotAdmin(t *testing.T) {
	dao := NewDAO("admin")

	dao.PendingApproval("user")
	dao.WhitelistUser("admin", "user")

	uassert.PanicsWithMessage(t, "only admins can demote members", func() { dao.RemoveAdminRank("user", "user") })
}

func Test_RemoveAdminRank_Failure_NotAdmin(t *testing.T) {
	dao := NewDAO("admin")

	dao.PendingApproval("user")
	dao.WhitelistUser("admin", "user")

	uassert.PanicsWithMessage(t, "member is not an admin", func() { dao.RemoveAdminRank("admin", "user") })
}

func Test_RemoveAdminRank_Failure_AutoDemotion(t *testing.T) {
	dao := NewDAO("admin")

	uassert.PanicsWithMessage(t, "cannot demote yourself", func() { dao.RemoveAdminRank("admin", "admin") })
}

// Test RemoveMember
func Test_RemoveMember_Success(t *testing.T) {
	dao := NewDAO("admin")

	dao.PendingApproval("user")
	dao.WhitelistUser("admin", "user")

	uassert.NotPanics(t, func() { dao.RemoveMember("admin", "user") })
}

func Test_RemoveMember_Failure_NotAdmin(t *testing.T) {
	dao := NewDAO("admin")

	dao.PendingApproval("user")
	dao.WhitelistUser("admin", "user")

	uassert.PanicsWithMessage(t, "must be an admin to remove a member", func() { dao.RemoveMember("user", "user") })
}

func Test_RemoveMember_Failure_NotMember(t *testing.T) {
	dao := NewDAO("admin")

	uassert.PanicsWithMessage(t, "not a member", func() { dao.RemoveMember("admin", "user") })
}

func Test_RemoveMember_Failure_AutoRemoval(t *testing.T) {
	dao := NewDAO("admin")

	uassert.PanicsWithMessage(t, "you cannot remove yourself", func() { dao.RemoveMember("admin", "admin") })
}

// Test LeaveDao
func Test_LeaveDao_Success(t *testing.T) {
	dao := NewDAO("admin")

	dao.PendingApproval("user")
	dao.WhitelistUser("admin", "user")

	uassert.NotPanics(t, func() { dao.LeaveDao("user") })
}

func Test_LeaveDao_Failure_NotMember(t *testing.T) {
	dao := NewDAO("admin")

	uassert.PanicsWithMessage(t, "not a member", func() { dao.LeaveDao("user") })
}

func Test_LeaveDao_Failure_LastAdmin(t *testing.T) {
	dao := NewDAO("admin")

	uassert.PanicsWithMessage(t, "you are the last admin: assign another admin before leaving", func() { dao.LeaveDao("admin") })
}
