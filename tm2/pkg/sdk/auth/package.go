package auth

import (
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/std"
)

var Package = amino.RegisterPackage(amino.NewPackage(
	"github.com/gnolang/gno/tm2/pkg/sdk/auth",
	"auth",
	amino.GetCallersDirname(),
).WithDependencies(
	std.Package,
).WithTypes(
	GenesisState{}, "GenesisState",
	Params{}, "Params",
	MsgCreateSession{}, "m_create_session",
	MsgRevokeSession{}, "m_revoke_session",
	MsgRevokeAllSessions{}, "m_revoke_all_sessions",
))
