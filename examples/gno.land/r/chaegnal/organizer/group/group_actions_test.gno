package group

import (
	"testing"

	"gno.land/p/demo/uassert"
	"gno.land/r/chaegnal/organizer"
)

// Test DeleteFriendGroup
func Test_DeleteFriendGroup_Success(t *testing.T) {
	testing.SetOriginCaller("Chaegnal")
	cross(organizer.CreateFriendGroup)("Chaegnal's group")

	uassert.NotPanics(t, func() { DeleteFriendGroup("1", "Chaegnal") })
}

func Test_DeleteFriendGroup_Failure(t *testing.T) {
	testing.SetOriginCaller("Chaegnal")
	cross(organizer.CreateFriendGroup)("Chaegnal's group")

	uassert.PanicsWithMessage(t, "only the owner can delete the group", func() { DeleteFriendGroup("2", "Ghhv") })
}

// Test SwitchOwner
func Test_SwitchOwner_Success(t *testing.T) {
	testing.SetOriginCaller("Chaegnal")
	cross(organizer.CreateFriendGroup)("Chaegnal's group")

	testing.SetOriginCaller("Ghhv")
	cross(organizer.AskJoinFriendGroup)(3)

	group := organizer.GetGroupFromId("3")
	group.DAO.WhitelistUser("Chaegnal", "Ghhv")

	testing.SetOriginCaller("Chaegnal")
	MakeAdmin("3", "Chaegnal", "Ghhv")

	uassert.NotPanics(t, func() { SwitchOwner("3", "Chaegnal", "Ghhv") })
}

func Test_SwitchOwner_NotOwner(t *testing.T) {
	testing.SetOriginCaller("Chaegnal")
	cross(organizer.CreateFriendGroup)("Chaegnal's group")

	testing.SetOriginCaller("Ghhv")
	uassert.PanicsWithMessage(t, "only the owner give ownership away", func() { SwitchOwner("4", "Ghhv", "Ghhv") })
}

func Test_SwitchOwner_NotAdmin(t *testing.T) {
	testing.SetOriginCaller("Chaegnal")
	cross(organizer.CreateFriendGroup)("Chaegnal's group")

	testing.SetOriginCaller("Ghhv")
	cross(organizer.AskJoinFriendGroup)(5)

	group := organizer.GetGroupFromId("5")
	group.DAO.WhitelistUser("Chaegnal", "Ghhv")

	testing.SetOriginCaller("Chaegnal")

	uassert.PanicsWithMessage(t, "the new owner must be an admin", func() { SwitchOwner("5", "Chaegnal", "Ghhv") })
}
