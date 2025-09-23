package testing_test

import "testing"

func Test_newRealm(t *testing.T) {
	ur := testing.NewUserRealm("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	const urExpected = "UserRealm{ g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5 }"
	if ur.String() != urExpected {
		t.Errorf("got %q want %q", ur.String(), urExpected)
	}
	cr := testing.NewCodeRealm("gno.land/r/gnoland/users")
	const crExpected = "CodeRealm{ g139h767xe0n2rtl2zntn8qjrkxrehfaaglep245, gno.land/r/gnoland/users }"
	if cr.String() != crExpected {
		t.Errorf("got %q want %q", cr.String(), crExpected)
	}
}
