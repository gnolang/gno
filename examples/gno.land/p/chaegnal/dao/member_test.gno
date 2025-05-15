package dao

import "testing"

// Test PendingApproval
func Test_PendingApproval_Success(t *testing.T) {
	defer func() {
		r := recover()
		if r != nil {
			t.Errorf("Unexpected panic: %v", r)
		}
	}()

	dao := NewDAO("admin")

	dao.PendingApproval("user")
}

func Test_PendingApproval_Failure_AlreadyAwaiting(t *testing.T) {
	defer func() {
		r := recover()
		if r != "already awaiting approval" {
			t.Errorf("Expected 'already awaiting approval' got '%v' panic", r)
		}
	}()

	dao := NewDAO("admin")

	dao.PendingApproval("user")
	dao.PendingApproval("user")
}

func Test_PendingApproval_Failure_AlreadyIn(t *testing.T) {
	defer func() {
		r := recover()
		if r != "already whitelisted" {
			t.Errorf("Expected 'already whitelisted' got '%v' panic", r)
		}
	}()

	dao := NewDAO("admin")

	dao.PendingApproval("admin")
}

// Test WhitelistUser
func Test_WhitelistUser_Success(t *testing.T) {
	defer func() {
		r := recover()
		if r != nil {
			t.Errorf("Unexpected panic: %v", r)
		}
	}()

	dao := NewDAO("admin")

	dao.PendingApproval("user")
	dao.WhitelistUser("admin", "user")

	member, ok := dao.IsMember("user")
	if !ok || member.Rank != memberRank {
		t.Errorf("Expected user to be whitelisted with member rank.")
	}
}

func Test_WhitelistUser_Failure_NotAdmin(t *testing.T) {
	defer func() {
		r := recover()
		if r != "only admins can whitelist members" {
			t.Errorf("Expected 'only admins can whitelist members' panic, got '%v' panic", r)
		}
	}()

	dao := NewDAO("admin")
	dao.PendingApproval("user")
	dao.WhitelistUser("user", "user")
}

func Test_WhitelistUser_Failure_NotAwaiting(t *testing.T) {
	defer func() {
		r := recover()
		if r != "this address is not awaiting approval" {
			t.Errorf("Expected 'this address is not awaiting approval' panic, got '%v' panic", r)
		}
	}()

	dao := NewDAO("admin")
	dao.WhitelistUser("admin", "user")
}

func Test_WhitelistUser_Failure_AlreadyIn(t *testing.T) {
	defer func() {
		r := recover()
		if r != "already whitelisted" {
			t.Errorf("Expected 'already whitelisted' panic, got '%v' panic", r)
		}
	}()

	dao := NewDAO("admin")
	dao.PendingApproval("user")
	dao.WhitelistUser("admin", "user")
	dao.PendingApproval("user")
}

// Test RefuseAwaitingMember
func Test_RefuseAwaitingMember_Success(t *testing.T) {
	defer func() {
		r := recover()
		if r != nil {
			t.Errorf("Unexpected panic: %v", r)
		}
	}()

	dao := NewDAO("admin")

	dao.PendingApproval("user")
	dao.RefuseAwaitingMember("admin", "user")
}

func Test_RefuseAwaitingMember_Failure(t *testing.T) {
	defer func() {
		r := recover()
		if r != "only admins can refuse members" {
			t.Errorf("Expected 'only admins can refuse members' panic, got '%v' panic", r)
		}
	}()

	dao := NewDAO("admin")

	dao.PendingApproval("user")
	dao.RefuseAwaitingMember("user", "user")
}

// Test SetAdmin
func Test_SetAdmin_Success(t *testing.T) {
	defer func() {
		r := recover()
		if r != nil {
			t.Errorf("Unexpected panic: %v", r)
		}
	}()

	dao := NewDAO("admin")

	dao.PendingApproval("user")
	dao.WhitelistUser("admin", "user")

	dao.SetAdmin("admin", "user")
}

func Test_SetAdmin_Failure_NotAdmin(t *testing.T) {
	defer func() {
		r := recover()
		if r != "only admins can promote members" {
			t.Errorf("Expected 'only admins can promote members' panic, got '%v' panic", r)
		}
	}()

	dao := NewDAO("admin")

	dao.PendingApproval("user")
	dao.WhitelistUser("admin", "user")

	dao.SetAdmin("user", "user")
}

// Test defaultAdmin
func Test_SetAdmin_Failure_NotMember(t *testing.T) {
	defer func() {
		r := recover()
		if r != "not a member" {
			t.Errorf("Expected 'not a member' panic, got '%v' panic", r)
		}
	}()

	dao := NewDAO("admin")

	dao.defaultAdmin("user")
}

func Test_SetAdmin_Failure_AlreadyAdmin(t *testing.T) {
	defer func() {
		r := recover()
		if r != "member is already an admin" {
			t.Errorf("Expected 'member is already an admin' panic, got '%v' panic", r)
		}
	}()

	dao := NewDAO("admin")

	dao.PendingApproval("user")
	dao.WhitelistUser("admin", "user")

	dao.defaultAdmin("user")
	dao.defaultAdmin("user")
}

// Test RemoveAdminRank
func Test_RemoveAdminRank_Success(t *testing.T) {
	defer func() {
		r := recover()
		if r != nil {
			t.Errorf("Unexpected panic: %v", r)
		}
	}()

	dao := NewDAO("admin")

	dao.PendingApproval("user")
	dao.WhitelistUser("admin", "user")
	dao.SetAdmin("admin", "user")

	dao.RemoveAdminRank("admin", "user")
}

func Test_RemoveAdminRank_Failure_CallerNotAdmin(t *testing.T) {
	defer func() {
		r := recover()
		if r != "only admins can demote members" {
			t.Errorf("Expected 'only admins can demote members' panic, got '%v' panic", r)
		}
	}()

	dao := NewDAO("admin")

	dao.PendingApproval("user")
	dao.WhitelistUser("admin", "user")

	dao.RemoveAdminRank("user", "user")
}

func Test_RemoveAdminRank_Failure_NotAdmin(t *testing.T) {
	defer func() {
		r := recover()
		if r != "member is not an admin" {
			t.Errorf("Expected 'member is not an admin' panic, got '%v' panic", r)
		}
	}()

	dao := NewDAO("admin")

	dao.PendingApproval("user")
	dao.WhitelistUser("admin", "user")

	dao.RemoveAdminRank("admin", "user")
}

func Test_RemoveAdminRank_Failure_AutoDemotion(t *testing.T) {
	defer func() {
		r := recover()
		if r != "cannot demote yourself" {
			t.Errorf("Expected 'cannot demote yourself' panic, got '%v' panic", r)
		}
	}()

	dao := NewDAO("admin")

	dao.RemoveAdminRank("admin", "admin")
}

// Test RemoveMember
func Test_RemoveMember_Success(t *testing.T) {
	defer func() {
		r := recover()
		if r != nil {
			t.Errorf("Unexpected panic: %v", r)
		}
	}()

	dao := NewDAO("admin")

	dao.PendingApproval("user")
	dao.WhitelistUser("admin", "user")

	dao.RemoveMember("admin", "user")
}

func Test_RemoveMember_Failure_NotAdmin(t *testing.T) {
	defer func() {
		r := recover()
		if r != "must be an admin to remove a member" {
			t.Errorf("Expected 'must be an admin to remove a member' panic, got '%v' panic", r)
		}
	}()

	dao := NewDAO("admin")

	dao.PendingApproval("user")
	dao.WhitelistUser("admin", "user")

	dao.RemoveMember("user", "user")
}

func Test_RemoveMember_Failure_NotMember(t *testing.T) {
	defer func() {
		r := recover()
		if r != "not a member" {
			t.Errorf("Expected 'not a member' panic, got '%v' panic", r)
		}
	}()

	dao := NewDAO("admin")

	dao.RemoveMember("admin", "user")
}

func Test_RemoveMember_Failure_AutoRemoval(t *testing.T) {
	defer func() {
		r := recover()
		if r != "you cannot remove yourself" {
			t.Errorf("Expected 'you cannot remove yourself' panic, got '%v' panic", r)
		}
	}()

	dao := NewDAO("admin")

	dao.RemoveMember("admin", "admin")
}

// Test LeaveDao
func Test_LeaveDao_Success(t *testing.T) {
	defer func() {
		r := recover()
		if r != nil {
			t.Errorf("Unexpected panic: %v", r)
		}
	}()

	dao := NewDAO("admin")

	dao.PendingApproval("user")
	dao.WhitelistUser("admin", "user")

	dao.LeaveDao("user")
}

func Test_LeaveDao_Failure_NotMember(t *testing.T) {
	defer func() {
		r := recover()
		if r != "not a member" {
			t.Errorf("Expected 'not a member' panic, got '%v' panic", r)
		}
	}()

	dao := NewDAO("admin")

	dao.LeaveDao("user")
}

func Test_LeaveDao_Failure_LastAdmin(t *testing.T) {
	defer func() {
		r := recover()
		if r != "you are the last admin: assign another admin before leaving" {
			t.Errorf("Expected 'you are the last admin: assign another admin before leaving' panic, got '%v' panic", r)
		}
	}()

	dao := NewDAO("admin")

	dao.LeaveDao("admin")
}
