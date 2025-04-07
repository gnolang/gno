package gnominers

import "std"

func assertCallerExists() *UserInfo {
	caller := std.PreviousRealm().Address()
	infos, ok := userSessions[caller.String()]
	if !ok {
		panic("user doesnt exists")
	}
	return infos
}