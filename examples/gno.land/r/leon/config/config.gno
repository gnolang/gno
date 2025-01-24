package config

import (
	"errors"
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ownable"
	"gno.land/p/demo/ufmt"
	"gno.land/p/moul/md"
)

var (
	configs = avl.NewTree()

	OwnableMain   = ownable.NewWithAddress("g125em6arxsnj49vx35f0n0z34putv5ty3376fg5")
	OwnableBackup = ownable.NewWithAddress("g125em6arxsnj49vx35f0n0z34putv5ty3376fg5")

	ErrUnauthorized = errors.New("leon's config: unauthorized")
)

func AddConfig(name, lines string) {
	if !IsAuthorized(std.PrevRealm().Addr()) {
		panic(ErrUnauthorized)
	}

	configs.Set(name, lines) // allows overwrite
}

func RemoveConfig(name string) {
	if !IsAuthorized(std.PrevRealm().Addr()) {
		panic(ErrUnauthorized)
	}

	if _, ok := configs.Remove(name); !ok {
		panic("no config with that name")
	}
}

func IsAuthorized(addr std.Address) bool {
	return addr == OwnableMain.Owner() || addr == OwnableBackup.Owner()
}

func Render(_ string) (out string) {
	out += md.H1("Leon's config Package")

	out += ufmt.Sprintf("Leon's main address: %s\n\n", OwnableMain.Owner().String())
	out += ufmt.Sprintf("Leon's backup address: %s\n\n", OwnableBackup.Owner().String())

	out += md.H2("Leon's configs")

	configs.Iterate("", "", func(key string, value interface{}) bool {
		out += ufmt.Sprintf("## %s\n\n", key)
		out += ufmt.Sprintf("```\n%s\n```", value.(string))

		return false
	})

	return out
}
