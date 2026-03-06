package realmpath_test

import (
	"chain/runtime"
	"net/url"
	"testing"

	"gno.land/p/moul/realmpath"
	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"
)

func TestExample(t *testing.T) {
	cd := runtime.ChainDomain()
	testing.SetRealm(testing.NewCodeRealm(cd + "/r/lorem/ipsum"))

	// initial parsing
	path := "hello/world?foo=bar&baz=foobar"
	req := realmpath.Parse(path)
	urequire.False(t, req == nil, "req should not be nil")
	uassert.Equal(t, req.Path, "hello/world")
	uassert.Equal(t, req.Query.Get("foo"), "bar")
	uassert.Equal(t, req.Query.Get("baz"), "foobar")
	uassert.Equal(t, req.String(), "/r/lorem/ipsum:hello/world?baz=foobar&foo=bar")

	// alter query
	req.Query.Set("hey", "salut")
	uassert.Equal(t, req.String(), "/r/lorem/ipsum:hello/world?baz=foobar&foo=bar&hey=salut")

	// alter path
	req.Path = "bye/ciao"
	uassert.Equal(t, req.String(), "/r/lorem/ipsum:bye/ciao?baz=foobar&foo=bar&hey=salut")
}

func TestParse(t *testing.T) {
	cd := runtime.ChainDomain()
	testing.SetRealm(testing.NewCodeRealm(cd + "/r/lorem/ipsum"))

	tests := []struct {
		rawPath        string
		realm_XXX      string // optional
		expectedPath   string
		expectedQuery  url.Values
		expectedString string
	}{
		{
			rawPath:      "hello/world?foo=bar&baz=foobar",
			expectedPath: "hello/world",
			expectedQuery: url.Values{
				"foo": []string{"bar"},
				"baz": []string{"foobar"},
			},
			expectedString: "/r/lorem/ipsum:hello/world?baz=foobar&foo=bar",
		},
		{
			rawPath:      "api/v1/resource?search=test&limit=10",
			expectedPath: "api/v1/resource",
			expectedQuery: url.Values{
				"search": []string{"test"},
				"limit":  []string{"10"},
			},
			expectedString: "/r/lorem/ipsum:api/v1/resource?limit=10&search=test",
		},
		{
			rawPath:        "singlepath",
			expectedPath:   "singlepath",
			expectedQuery:  url.Values{},
			expectedString: "/r/lorem/ipsum:singlepath",
		},
		{
			rawPath:        "path/with/trailing/slash/",
			expectedPath:   "path/with/trailing/slash/",
			expectedQuery:  url.Values{},
			expectedString: "/r/lorem/ipsum:path/with/trailing/slash",
		},
		{
			rawPath:        "emptyquery?",
			expectedPath:   "emptyquery",
			expectedQuery:  url.Values{},
			expectedString: "/r/lorem/ipsum:emptyquery",
		},
		{
			rawPath:      "path/with/special/characters/?key=val%20ue&anotherKey=with%21special%23chars",
			expectedPath: "path/with/special/characters/",
			expectedQuery: url.Values{
				"key":        []string{"val ue"},
				"anotherKey": []string{"with!special#chars"},
			},
			expectedString: "/r/lorem/ipsum:path/with/special/characters?anotherKey=with%21special%23chars&key=val+ue",
		},
		{
			rawPath:      "path/with/empty/key?keyEmpty&=valueEmpty",
			expectedPath: "path/with/empty/key",
			expectedQuery: url.Values{
				"keyEmpty": []string{""},
				"":         []string{"valueEmpty"},
			},
			expectedString: "/r/lorem/ipsum:path/with/empty/key?=valueEmpty&keyEmpty=",
		},
		{
			rawPath:      "path/with/multiple/empty/keys?=empty1&=empty2",
			expectedPath: "path/with/multiple/empty/keys",
			expectedQuery: url.Values{
				"": []string{"empty1", "empty2"},
			},
			expectedString: "/r/lorem/ipsum:path/with/multiple/empty/keys?=empty1&=empty2",
		},
		{
			rawPath:      "path/with/percent-encoded/%20space?query=hello%20world",
			expectedPath: "path/with/percent-encoded/%20space", // XXX: should we decode?
			expectedQuery: url.Values{
				"query": []string{"hello world"},
			},
			expectedString: "/r/lorem/ipsum:path/with/percent-encoded/%20space?query=hello+world",
		},
		{
			rawPath:      "path/with/very/long/query?key1=value1&key2=value2&key3=value3&key4=value4&key5=value5&key6=value6",
			expectedPath: "path/with/very/long/query",
			expectedQuery: url.Values{
				"key1": []string{"value1"},
				"key2": []string{"value2"},
				"key3": []string{"value3"},
				"key4": []string{"value4"},
				"key5": []string{"value5"},
				"key6": []string{"value6"},
			},
			expectedString: "/r/lorem/ipsum:path/with/very/long/query?key1=value1&key2=value2&key3=value3&key4=value4&key5=value5&key6=value6",
		},
		{
			rawPath:      "custom/realm?foo=bar&baz=foobar",
			realm_XXX:    cd + "/r/foo/bar",
			expectedPath: "custom/realm",
			expectedQuery: url.Values{
				"foo": []string{"bar"},
				"baz": []string{"foobar"},
			},
			expectedString: "/r/foo/bar:custom/realm?baz=foobar&foo=bar",
		},
	}

	for _, tt := range tests {
		t.Run(tt.rawPath, func(t *testing.T) {
			req := realmpath.Parse(tt.rawPath)
			req.Realm = tt.realm_XXX // set optional realm
			urequire.False(t, req == nil, "req should not be nil")
			uassert.Equal(t, req.Path, tt.expectedPath)
			urequire.Equal(t, len(req.Query), len(tt.expectedQuery))
			uassert.Equal(t, req.Query.Encode(), tt.expectedQuery.Encode())
			// XXX: uassert.Equal(t, req.Query, tt.expectedQuery)
			uassert.Equal(t, req.String(), tt.expectedString)
		})
	}
}
