package gnominers

import (
	"std"
	"strconv"
)

func assertCallerExists() *UserSession {
	caller := std.PreviousRealm().Address()
	infos, ok := userSessions[caller.String()]
	if !ok {
		panic("user doesnt exists")
	}
	return infos
}

func mustAtoi(s string) int {
	x, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return x
}
