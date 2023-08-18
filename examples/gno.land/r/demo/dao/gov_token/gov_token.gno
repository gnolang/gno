package gov_token

import (
	"gno.land/p/demo/gor/snapshot_token"
	"std"
)

// realm keep snapshot tokens as a gov token, (snapshot) native coin, grc series coin, etc

var GovToken *snapshot_token.SnapshotToken

func init() {
	GovToken = snapshot_token.NewSnapShotToken()
	println(GovToken.Echo())
}

func GetToken(acc std.Address) uint64 {
	return GovToken.GetToken(acc)
}
