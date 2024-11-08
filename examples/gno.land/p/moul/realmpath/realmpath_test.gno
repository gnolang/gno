package realmpath_test

import (
	"std"
	"testing"

	"gno.land/p/demo/uassert"
	"gno.land/p/demo/urequire"
	"gno.land/p/moul/realmpath"
)

func TestExample(t *testing.T) {
	// initial parsing
	path := "hello/world?foo=bar&baz=foobar"
	req := realmpath.Parse(path)
	urequire.False(t, req == nil, "req should not be nil")
	uassert.Equal(t, req.Path, "hello/world")
	uassert.Equal(t, req.Query.Get("foo"), "bar")
	uassert.Equal(t, req.Query.Get("baz"), "foobar")
	std.TestSetRealm(std.NewCodeRealm("gno.land/r/lorem/ipsum"))
	uassert.Equal(t, req.String(), "/r/lorem/ipsum:hello/world?baz=foobar&foo=bar")

	// alter query
	req.Query.Set("hey", "salut")
	uassert.Equal(t, req.String(), "/r/lorem/ipsum:hello/world?baz=foobar&foo=bar&hey=salut")

	// alter path
	req.Path = "bye/ciao"
	uassert.Equal(t, req.String(), "/r/lorem/ipsum:bye/ciao?baz=foobar&foo=bar&hey=salut")
}
