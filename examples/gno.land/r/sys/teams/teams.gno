package teams

import (
	"errors"
	"std"
	"strings"

	"gno.land/p/demo/avl"
)

type Team struct {
	Admins  []std.Address
	Editors []std.Address
}

var (
	ErrNotHomeCaller  = errors.New("must be called by r/<team>/home userland logic")
	ErrInvalidMembers = errors.New("invalid members argument")
)

type Role uint

const (
	RoleUnset  = ""
	RoleAdmin  = "admin"  // editor + right to upgrade r/<team>/home
	RoleEditor = "editor" // can add package under {p,r}/<team>/...
	// XXX: RoleReadonly = "readonly"?
)

var teams avl.Tree // team -> []users

// TODO: doc
func SetMembers(attrs ...string) {
	// XXX: panic if members list is too big for now.

	attrsLen := len(attrs)
	if attrsLen%2 != 0 {
		panic(ErrInvalidMembers)
	}

	prev := std.PrevRealm()
	if prev.IsUser() {
		panic(ErrNotHomeCaller)
	}

	prevPath := prev.PkgPath()
	parts := strings.Split(prevPath, "/")
	if len(parts) != 4 || !strings.HasPrefix(prevPath, "gno.land/r/") || !strings.HasSuffix(prevPath, "/home") {
		panic(ErrNotHomeCaller)
	}
	teamName := parts[2]

	team := Team{
		Admins:  []std.Address{},
		Editors: []std.Address{},
	}
	for i := 0; i < attrsLen-1; i += 2 {
		addr := std.Address(attrs[i])
		role := attrs[i+1]
		if role == RoleAdmin {
			team.Admins = append(team.Admins, addr)
		}
		if role == RoleEditor {
			team.Editors = append(team.Editors, addr)
		}
	}

	teams.Set(teamName, &team)
}

func GetAdmins(teamName string) []std.Address {
	t, ok := teams.Get(teamName)
	if !ok {
		return nil
	}
	team := t.(*Team)
	return team.Admins
}

func GetEditors(teamName string) []std.Address {
	t, ok := teams.Get(teamName)
	if !ok {
		return nil
	}
	team := t.(*Team)
	return append(team.Admins, team.Editors...)
}

// XXX: func GetUserTeams(addr std.Address)
