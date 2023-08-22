package gov_token

import (
	"gno.land/p/demo/gor/snapshot_token"
	"std"
)

// realm keep snapshot tokens as a gov token, (snapshot) native coin, grc series coin, etc
// TODO: might not need all these intermediate representation, since weight of member has been snapshotted

var GovToken *snapshot_token.SnapshotToken

func init() {
	GovToken = snapshot_token.NewSnapShotToken()
}

func GetAmount(acc std.Address) uint64 {
	return GovToken.GetAmount(acc)
}

func GetTotalSupply() uint64 {
	return GovToken.GetTotalSupply()
}

func TransferVotingPower(from, to std.Address, amount uint64) {
	GovToken.TransferVotingUnits(from, to, amount)
}
