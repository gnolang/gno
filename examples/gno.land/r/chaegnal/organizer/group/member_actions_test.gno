package group

import (
	"testing"

	"gno.land/p/demo/uassert"
	"gno.land/r/chaegnal/organizer"
)

// RemoveAdminRank
func Test_RemoveAdminRank_Success(t *testing.T) {
	testing.SetOriginCaller("Chaegnal")
	cross(organizer.CreateFriendGroup)("Chaegnal's group")

	testing.SetOriginCaller("Ghhv")
	cross(organizer.AskJoinFriendGroup)(7)

	group := organizer.GetGroupFromId("7")
	group.DAO.WhitelistUser("Chaegnal", "Ghhv")

	testing.SetOriginCaller("Chaegnal")
	MakeAdmin("7", "Chaegnal", "Ghhv")

	uassert.NotPanics(t, func() { RemoveAdminRank("7", "Chaegnal", "Ghhv") })
}

func Test_RemoveAdminRank_Failure(t *testing.T) {
	testing.SetOriginCaller("Chaegnal")
	cross(organizer.CreateFriendGroup)("Chaegnal's group")

	testing.SetOriginCaller("Ghhv")
	cross(organizer.AskJoinFriendGroup)(8)

	group := organizer.GetGroupFromId("8")
	group.DAO.WhitelistUser("Chaegnal", "Ghhv")

	testing.SetOriginCaller("Chaegnal")
	MakeAdmin("8", "Chaegnal", "Ghhv")

	testing.SetOriginCaller("Ghhv")

	uassert.PanicsWithMessage(t, "you cannot undo admin rights of the owner", func() { RemoveAdminRank("8", "Ghhv", "Chaegnal") })
}
