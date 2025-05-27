package group

import (
	"testing"

	"gno.land/r/chaegnal/organizer"
)

// RemoveAdminRank
func Test_RemoveAdminRank_Success(t *testing.T) {
	defer func() {
		r := recover()
		if r != nil {
			t.Errorf("Unexpected panic: %v", r)
		}
	}()

	testing.SetOriginCaller("Chaegnal")
	cross(organizer.CreateFriendGroup)("Chaegnal's group")

	testing.SetOriginCaller("Ghhv")
	cross(organizer.AskJoinFriendGroup)(7)

	group := organizer.GetGroupFromId("7")
	group.DAO.WhitelistUser("Chaegnal", "Ghhv")

	testing.SetOriginCaller("Chaegnal")
	MakeAdmin("7", "Chaegnal", "Ghhv")

	RemoveAdminRank("7", "Chaegnal", "Ghhv")
}

func Test_RemoveAdminRank_Failure(t *testing.T) {
	defer func() {
		r := recover()
		if r != "you cannot undo admin rights of the owner" {
			t.Errorf("Expected 'you cannot undo admin rights of the owner' panic, got '%v' panic", r)
		}
	}()

	testing.SetOriginCaller("Chaegnal")
	cross(organizer.CreateFriendGroup)("Chaegnal's group")

	testing.SetOriginCaller("Ghhv")
	cross(organizer.AskJoinFriendGroup)(8)

	group := organizer.GetGroupFromId("8")
	group.DAO.WhitelistUser("Chaegnal", "Ghhv")

	testing.SetOriginCaller("Chaegnal")
	MakeAdmin("8", "Chaegnal", "Ghhv")

	testing.SetOriginCaller("Ghhv")
	RemoveAdminRank("8", "Ghhv", "Chaegnal")
}
