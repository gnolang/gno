package names

import (
	"std"
	"testing"
)

func TestRegisterNamespace(t *testing.T) {
	std.TestSetOrigCaller(std.Address("g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq"))
	creator := std.GetOrigCaller()

	n := "test2"
	Register(n)
	s := getSpace(n)
	assertTrue(t, containsAddress(s.Admins, creator))
}

func TestAdminFuncs(t *testing.T) {
	std.TestSetOrigCaller(std.Address("g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq"))
	creator := std.GetOrigCaller()

	test1 := std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	test2 := std.Address("g1r6casl322p5adkmmjjkjea404e7f6w29y6gzwg")

	ns := "test3"

	Register(ns)
	s := getSpace(ns)

	AddAdmin(ns, test1)
	AddAdmin(ns, test2)
	assertTrue(t, containsAddress(s.Admins, test1))
	assertTrue(t, containsAddress(s.Admins, test2))

	RemoveAdmin(ns, test1)
	assertFalse(t, containsAddress(s.Admins, test1))
	assertTrue(t, containsAddress(s.Admins, test2))

	AddEditor(ns, test1)
	assertTrue(t, containsAddress(s.Editors, test1))

	RemoveEditor(ns, test1)
	assertTrue(t, !containsAddress(s.Editors, test1))
}

func TestSpaceMethods(t *testing.T) {
	std.TestSetOrigCaller(std.Address("g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq"))
	creator := std.GetOrigCaller()

	test1 := std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	test2 := std.Address("g1r6casl322p5adkmmjjkjea404e7f6w29y6gzwg")

	ns := "test4"
	Register(ns)
	s := getSpace(ns)

	assertTrue(t, s.isAdmin(creator))
	assertTrue(t, s.removeAdmin(creator) != nil)

	s.addAdmin(test1)
	assertTrue(t, s.removeAdmin(test1) == nil)
}

func assertTrue(t *testing.T, b bool) {
	if !b {
		t.Fail()
	}
}

func assertFalse(t *testing.T, b bool) {
	if b {
		t.Fail()
	}
}
