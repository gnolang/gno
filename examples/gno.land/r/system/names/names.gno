// The realm r/system/names is used to manage namespaces on gno.land.
package names

import (
	"errors"
	"regexp"
	"std"

	"gno.land/p/demo/avl"
)

// "AddPkg" will check if r/system/names exists. If yes, it will
// inspect the realm's state and use the following variable to
// determine if an address can publish a package or not.
var namespaces avl.Tree // name(string) -> Space

// TODO: more accurate.
var reNamespace = regexp.MustCompile(`^[a-z][a-z0-9_]{2,30}$`)

type Space struct {
	Admins  []std.Address
	Editors []std.Address
	InPause bool
}

func (s *Space) isAdmin(addr std.Address) bool {
	return containsAddress(s.Admins, addr)
}

func (s *Space) addAdmin(newAdmin std.Address) {
	if !containsAddress(s.Admins, newAdmin) {
		s.Admins = append(s.Admins, newAdmin)
	}
}

func (s *Space) removeAdmin(admin std.Address) error {
	if len(s.Admins) == 1 {
		return errors.New("namespace at least needs one admin")
	}
	admins := removeAddress(s.Admins, admin)
	s.Admins = admins
	return nil
}

func (s *Space) addEditor(newEditor std.Address) {
	if !containsAddress(s.Editors, newEditor) {
		s.Editors = append(s.Editors, newEditor)
	}
}

func (s *Space) removeEditor(editor std.Address) error {
	editors := removeAddress(s.Editors, editor)
	s.Editors = editors
	return nil
}

func (s *Space) hasPerm(caller std.Address) bool {
	if s.InPause {
		return false
	}

	if containsAddress(s.Admins, caller) {
		return true
	}

	if containsAddress(s.Editors, caller) {
		return true
	}

	return false
}

func validateNamespace(namespace string) error {
	if !reNamespace.MatchString(namespace) {
		return errors.New("invalid namespace")
	}
	return nil
}
