package group

import (
	"testing"

	"gno.land/p/demo/uassert"
	"gno.land/r/chaegnal/organizer"
)

// Test DeleteFriendGroup
func Test_DeleteFriendGroup_Success(t *testing.T) {
	testing.SetOriginCaller("Chaegnal")
	organizer.CreateFriendGroup(cross, "Chaegnal's group")

	uassert.NotPanics(t, func() { DeleteFriendGroup(cross, "1", "Chaegnal") })
}

func Test_DeleteFriendGroup_Failure(t *testing.T) {
	testing.SetOriginCaller("Chaegnal")
	organizer.CreateFriendGroup(cross, "Chaegnal's group")

	uassert.PanicsWithMessage(t, "only the owner can delete the group", func() { DeleteFriendGroup(cross, "2", "Ghhv") })
}

// Test SwitchOwner
func Test_SwitchOwner_Success(t *testing.T) {
	testing.SetOriginCaller("Chaegnal")
	organizer.CreateFriendGroup(cross, "Chaegnal's group")

	testing.SetOriginCaller("Ghhv")
	organizer.AskJoinFriendGroup(cross, 3)

	group := organizer.GetGroupFromId("3")
	group.DAO.WhitelistUser("Chaegnal", "Ghhv")

	testing.SetOriginCaller("Chaegnal")
	MakeAdmin(cross, "3", "Chaegnal", "Ghhv")

	uassert.NotPanics(t, func() { SwitchOwner(cross, "3", "Chaegnal", "Ghhv") })
}

func Test_SwitchOwner_NotOwner(t *testing.T) {
	testing.SetOriginCaller("Chaegnal")
	organizer.CreateFriendGroup(cross, "Chaegnal's group")

	testing.SetOriginCaller("Ghhv")
	uassert.PanicsWithMessage(t, "only the owner give ownership away", func() { SwitchOwner(cross, "4", "Ghhv", "Ghhv") })
}

func Test_SwitchOwner_NotAdmin(t *testing.T) {
	testing.SetOriginCaller("Chaegnal")
	organizer.CreateFriendGroup(cross, "Chaegnal's group")

	testing.SetOriginCaller("Ghhv")
	organizer.AskJoinFriendGroup(cross, 5)

	group := organizer.GetGroupFromId("5")
	group.DAO.WhitelistUser("Chaegnal", "Ghhv")

	testing.SetOriginCaller("Chaegnal")

	uassert.PanicsWithMessage(t, "the new owner must be an admin", func() { SwitchOwner(cross, "5", "Chaegnal", "Ghhv") })
}
