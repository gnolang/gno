package commondao

import (
	"std"
	"strings"

	"gno.land/p/nt/ufmt"
)

func currentRealmPath() string {
	return strings.TrimPrefix(string(realmLink), std.ChainDomain())
}

func daoURL(daoID uint64) string {
	path := currentRealmPath()
	return ufmt.Sprintf("%s:%d", path, daoID)
}

func settingsURL(daoID uint64) string {
	path := currentRealmPath()
	return ufmt.Sprintf("%s:%d/settings", path, daoID)
}

func daoProposalsURL(daoID uint64) string {
	path := currentRealmPath()
	return ufmt.Sprintf("%s:%d/proposals", path, daoID)
}

func proposalURL(daoID, proposalID uint64) string {
	path := currentRealmPath()
	return ufmt.Sprintf("%s:%d/proposals/%d", path, daoID, proposalID)
}

func voteURL(daoID, proposalID uint64, addr std.Address) string {
	path := currentRealmPath()
	return ufmt.Sprintf("%s:%d/proposals/%d/vote/%s", path, daoID, proposalID, addr)
}
