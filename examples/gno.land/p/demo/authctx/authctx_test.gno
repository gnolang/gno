package authctx

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
)

func TestNewOrigAuthContext(t *testing.T) {
	var test1 std.Address = testutils.TestAddress("test1")
	std.TestSetOrigCaller(test1)

	auth := NewOrigAuthContext()

	assertSameAddresses(t, test1, auth.Addr())
	assertSameStrings(t, "orig:"+string(test1), auth.String())
	if !auth.IsAuthorized("foo") {
		t.Errorf("should be authorized")
	}
}

func TestNewPrevAuthContext(t *testing.T) {
	var test1 std.Address = testutils.TestAddress("test1")
	std.TestSetOrigCaller(test1)

	auth := NewPrevAuthContext()

	assertSameAddresses(t, test1, auth.Addr())
	assertSameStrings(t, "prev:"+string(test1), auth.String())
	if !auth.IsAuthorized("foo") {
		t.Errorf("should be authorized")
	}
}

func TestPackage(t *testing.T) {
	var (
		test1 = testutils.TestAddress("test1")
		test2 = testutils.TestAddress("test2")
	)

	tt := []struct {
		name          string
		init          func() AuthContext
		expNilAddr    bool
		expAddr       std.Address
		expAuthorized bool
	}{
		{
			name: "nil",
			init: func() AuthContext {
				var ctx AuthContext = nil
				return ctx
			},
			expNilAddr: true,
		},
		{
			name: "orig",
			init: func() AuthContext {
				std.TestSetOrigCaller(test1)
				return NewOrigAuthContext()
			},
			expAddr:       test1,
			expAuthorized: true,
		},
		{
			name: "orig-scoped-authorized",
			init: func() AuthContext {
				std.TestSetOrigCaller(test1)
				ctx := NewOrigAuthContext()
				scoped := NewScopedAuthContext(ctx, func(action string) bool { return true })
				return scoped
			},
			expAddr:       test1,
			expAuthorized: true,
		},
		{
			name: "orig-scoped-unauthorized",
			init: func() AuthContext {
				std.TestSetOrigCaller(test1)
				ctx := NewOrigAuthContext()
				scoped := NewScopedAuthContext(ctx, func(action string) bool { return false })
				return scoped
			},
			expAddr:       test1,
			expAuthorized: false,
		},
		{
			name: "delegated-orig-authorized",
			init: func() AuthContext {
				var db DelegationDB
				std.TestSetOrigCaller(test1)
				db.Approve(test2, AuthorizeAll())
				std.TestSetOrigCaller(test2)
				return db.NewOrigDelegatedAuthContext(test1)
			},
			expAddr:       test1,
			expAuthorized: true,
		},
		{
			name: "delegated-orig-authorized",
			init: func() AuthContext {
				var db DelegationDB
				std.TestSetOrigCaller(test1)
				db.Approve(test2, AuthorizeAll())
				std.TestSetOrigCaller(test2)
				return db.NewOrigDelegatedAuthContext(test1)
			},
			expAddr:       test1,
			expAuthorized: true,
		},
		{
			name: "delegated-orig-unauthorized",
			init: func() AuthContext {
				var db DelegationDB
				std.TestSetOrigCaller(test1)
				db.Approve(test2, AuthorizeNone())
				std.TestSetOrigCaller(test2)
				return db.NewOrigDelegatedAuthContext(test1)
			},
			expAddr:       test1,
			expAuthorized: false,
		},
		{
			name: "delegated-orig-conditional-authorized",
			init: func() AuthContext {
				var db DelegationDB
				std.TestSetOrigCaller(test1)
				db.Approve(test2, AuthorizeActions("foo", "bar", "foobar"))
				std.TestSetOrigCaller(test2)
				return db.NewOrigDelegatedAuthContext(test1)
			},
			expAddr:       test1,
			expAuthorized: true,
		},
		{
			name: "delegated-orig-custom-authfn",
			init: func() AuthContext {
				var db DelegationDB
				std.TestSetOrigCaller(test1)
				var customAuthFn AuthFunc = func(action string) bool { return action == "foobar" }
				db.Approve(test2, customAuthFn)
				std.TestSetOrigCaller(test2)
				return db.NewOrigDelegatedAuthContext(test1)
			},
			expAddr:       test1,
			expAuthorized: true,
		},
		{
			name: "delegated-non-preapproved",
			init: func() AuthContext {
				var db DelegationDB
				std.TestSetOrigCaller(test2)
				return db.NewOrigDelegatedAuthContext(test1)
			},
			expNilAddr: true,
		},

		// TODO: {name: "prev"} // depends on TestSetPrevRealm
		// TODO: {name: "owner-is-a-realm"},
		// TODO: {name: "non-address-owner"},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ctx := tc.init()
			if tc.expAddr == nil && ctx != nil {
				t.Errorf("expected AuthContext to be nil, but is %q", ctx)
			}
			if ctx == nil {
				return
			}
			if tc.expAddr != ctx.Addr() {
				t.Errorf("invalid addr, expected %q, got %q", tc.expAddr, ctx.Addr())
			}
			isAuthorized := ctx.IsAuthorized("foobar")
			if isAuthorized && !tc.expAuthorized {
				t.Errorf("expected unauthorized but was authorized")
			}
			if !isAuthorized && tc.expAuthorized {
				t.Errorf("expected authorized but was unauthorized")
			}
		})
	}
}

func assertSameStrings(t *testing.T, expected, got string) {
	t.Helper()
	if expected != got {
		t.Errorf("expected %q, got %q.", expected, got)
	}
}

func assertSameAddresses(t *testing.T, expected, got std.Address) {
	t.Helper()
	if expected != got {
		t.Errorf("expected %q, got %q.", expected, got)
	}
}
