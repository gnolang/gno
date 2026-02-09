package weburl

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseGnoURL(t *testing.T) {
	testCases := []struct {
		Name          string
		Input         string
		Expected      *GnoURL
		Err           error
		IsInvalidPath bool
	}{
		{
			Name:  "simple",
			Input: "https://gno.land/r/simple/test",
			Expected: &GnoURL{
				Domain:   "gno.land",
				Path:     "/r/simple/test",
				WebQuery: url.Values{},
				Query:    url.Values{},
			},
		},

		{
			Name:  "file",
			Input: "https://gno.land/r/simple/test/encode.gno",
			Expected: &GnoURL{
				Domain:   "gno.land",
				Path:     "/r/simple/test",
				WebQuery: url.Values{},
				Query:    url.Values{},
				File:     "encode.gno",
			},
		},

		{
			Name:  "complex file path",
			Input: "https://gno.land/r/simple/test///...gno",
			Expected: &GnoURL{
				Domain:   "gno.land",
				Path:     "/r/simple/test//",
				WebQuery: url.Values{},
				Query:    url.Values{},
				File:     "...gno",
			},
		},

		{
			Name:  "webquery + query",
			Input: "https://gno.land/r/demo/foo$help&func=Bar&name=Baz",
			Expected: &GnoURL{
				Path: "/r/demo/foo",
				Args: "",
				WebQuery: url.Values{
					"help": []string{""},
					"func": []string{"Bar"},
					"name": []string{"Baz"},
				},
				Query:  url.Values{},
				Domain: "gno.land",
			},
		},

		{
			Name:  "path args + webquery",
			Input: "https://gno.land/r/demo/foo:example$tz=Europe/Paris",
			Expected: &GnoURL{
				Path: "/r/demo/foo",
				Args: "example",
				WebQuery: url.Values{
					"tz": []string{"Europe/Paris"},
				},
				Query:  url.Values{},
				Domain: "gno.land",
			},
		},

		{
			Name:  "path args + webquery + query",
			Input: "https://gno.land/r/demo/foo:example$tz=Europe/Paris?hello=42",
			Expected: &GnoURL{
				Path: "/r/demo/foo",
				Args: "example",
				WebQuery: url.Values{
					"tz": []string{"Europe/Paris"},
				},
				Query: url.Values{
					"hello": []string{"42"},
				},
				Domain: "gno.land",
			},
		},

		{
			Name:  "webquery inside query",
			Input: "https://gno.land/r/demo/foo:example?value=42$tz=Europe/Paris",
			Expected: &GnoURL{
				Path:     "/r/demo/foo",
				Args:     "example",
				WebQuery: url.Values{},
				Query: url.Values{
					"value": []string{"42$tz=Europe/Paris"},
				},
				Domain: "gno.land",
			},
		},

		{
			Name:  "webquery escaped $",
			Input: "https://gno.land/r/demo/foo:example%24hello=43$hello=42",
			Expected: &GnoURL{
				Path: "/r/demo/foo",
				Args: "example$hello=43",
				WebQuery: url.Values{
					"hello": []string{"42"},
				},
				Query:  url.Values{},
				Domain: "gno.land",
			},
		},

		{
			Name:  "unknown path kind",
			Input: "https://gno.land/x/demo/foo",
			Expected: &GnoURL{
				Path:     "/x/demo/foo",
				Args:     "",
				WebQuery: url.Values{},
				Query:    url.Values{},
				Domain:   "gno.land",
			},
		},

		{
			Name:  "empty path",
			Input: "https://gno.land/r/",
			Expected: &GnoURL{
				Path:     "/r/",
				Args:     "",
				WebQuery: url.Values{},
				Query:    url.Values{},
				Domain:   "gno.land",
			},
		},

		{
			Name:  "complex query",
			Input: "https://gno.land/r/demo/foo$help?func=Bar&name=Baz&age=30",
			Expected: &GnoURL{
				Path: "/r/demo/foo",
				Args: "",
				WebQuery: url.Values{
					"help": []string{""},
				},
				Query: url.Values{
					"func": []string{"Bar"},
					"name": []string{"Baz"},
					"age":  []string{"30"},
				},
				Domain: "gno.land",
			},
		},

		{
			Name:  "multiple web queries",
			Input: "https://gno.land/r/demo/foo$help&func=Bar$test=123",
			Expected: &GnoURL{
				Path: "/r/demo/foo",
				Args: "",
				WebQuery: url.Values{
					"help": []string{""},
					"func": []string{"Bar$test=123"},
				},
				Query:  url.Values{},
				Domain: "gno.land",
			},
		},

		{
			Name:  "webquery-args-webquery",
			Input: "https://gno.land/r/demo/aaa$bbb:CCC&DDD$EEE",
			Err:   ErrURLInvalidPath, // `/r/demo/aaa$bbb` is an invalid path
		},

		{
			Name:  "args-webquery-args",
			Input: "https://gno.land/r/demo/aaa:BBB$CCC&DDD:EEE",
			Expected: &GnoURL{
				Domain: "gno.land",
				Path:   "/r/demo/aaa",
				Args:   "BBB",
				WebQuery: url.Values{
					"CCC":     []string{""},
					"DDD:EEE": []string{""},
				},
				Query: url.Values{},
			},
		},

		{
			Name:  "escaped characters in args",
			Input: "https://gno.land/r/demo/foo:example%20with%20spaces$tz=Europe/Paris",
			Expected: &GnoURL{
				Path: "/r/demo/foo",
				Args: "example with spaces",
				WebQuery: url.Values{
					"tz": []string{"Europe/Paris"},
				},
				Query:  url.Values{},
				Domain: "gno.land",
			},
		},

		{
			Name:  "file in path + args + query",
			Input: "https://gno.land/r/demo/foo/render.gno:example$tz=Europe/Paris",
			Expected: &GnoURL{
				Path: "/r/demo/foo",
				File: "render.gno",
				Args: "example",
				WebQuery: url.Values{
					"tz": []string{"Europe/Paris"},
				},
				Query:  url.Values{},
				Domain: "gno.land",
			},
		},

		{
			Name:  "no extension file",
			Input: "https://gno.land/r/demo/lIcEnSe",
			Expected: &GnoURL{
				Path:     "/r/demo",
				File:     "lIcEnSe",
				Args:     "",
				WebQuery: url.Values{},
				Query:    url.Values{},
				Domain:   "gno.land",
			},
		},

		{
			Name:     "invalid path",
			Input:    "https://gno.land/r/dem)o:$?",
			Expected: nil,
			Err:      ErrURLInvalidPath,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Logf("testing input: %q", tc.Input)

			u, err := url.Parse(tc.Input)
			require.NoError(t, err)

			result, err := ParseFromURL(u)
			if tc.Err == nil {
				require.NoError(t, err)
				t.Logf("encoded web path: %q", result.EncodeWebURL())
			} else {
				require.Error(t, err)
				require.ErrorIs(t, err, tc.Err)
			}

			assert.Equal(t, tc.Expected, result)
		})
	}
}

func TestIsValidPath(t *testing.T) {
	testCases := []struct {
		Path  string
		Valid bool
	}{
		{Path: "/", Valid: true},
		{Path: "/r/valid", Valid: true},
		{Path: "/p/abc_123", Valid: true},
		{Path: "/r/demo/users/", Valid: true},
		{Path: "/R/invalid", Valid: false},
		{Path: "/r/invalid@path", Valid: false},
		{Path: "r/invalid", Valid: false},
		{Path: "", Valid: false},
		{Path: "/r/valid/path_with/underscores", Valid: true},
		{Path: "/r/", Valid: true},
		{Path: "/r/with space", Valid: false},
		{Path: "/r/hyphen-invalid", Valid: false},
	}

	for _, tc := range testCases {
		t.Run(tc.Path, func(t *testing.T) {
			gnoURL := GnoURL{Path: tc.Path}
			assert.Equal(t, tc.Valid, gnoURL.IsValidPath())
		})
	}
}

func TestNamespace(t *testing.T) {
	testCases := []struct {
		Path     string
		Expected string
	}{
		{Path: "/r/test", Expected: "test"},
		{Path: "/r/test/foo", Expected: "test"},
		{Path: "/p/another", Expected: "another"},
		{Path: "/r/123invalid", Expected: ""},
		{Path: "/r/TEST", Expected: ""},
		{Path: "/x/ns", Expected: "ns"},
		{Path: "/r/a", Expected: "a"},
		{Path: "/r/a1", Expected: "a1"},
		{Path: "/r/a_b/c", Expected: "a_b"},
		{Path: "/invalidpath", Expected: ""},
		{Path: "/r/", Expected: ""},
		{Path: "/r/a-b/c", Expected: ""},
		{Path: "/r/valid-ns", Expected: ""},
	}

	for _, tc := range testCases {
		t.Run(tc.Path, func(t *testing.T) {
			gnoURL := GnoURL{Path: tc.Path}
			assert.Equal(t, tc.Expected, gnoURL.Namespace())
		})
	}
}

func TestEncode(t *testing.T) {
	testCases := []struct {
		Name        string
		GnoURL      GnoURL
		EncodeFlags EncodeFlag
		Expected    string
	}{
		{
			Name: "Encode domain",
			GnoURL: GnoURL{
				Domain: "gno.land",
				Path:   "/r/demo/foo",
			},
			EncodeFlags: EncodeDomain,
			Expected:    "gno.land",
		},

		{
			Name: "Encode web query with NoEscape",
			GnoURL: GnoURL{
				Domain: "gno.land",
				Path:   "/r/demo/foo",
				WebQuery: url.Values{
					"help":  []string{""},
					"fun$c": []string{"B$ ar"},
				},
			},
			EncodeFlags: EncodeWebQuery | EncodeNoEscape,
			// (still encoded; we always url-encode arguments after $ and ?)
			Expected: "$fun%24c=B%24+ar&help",
		},

		{
			Name: "encode domain and path",
			GnoURL: GnoURL{
				Domain: "gno.land",
				Path:   "/r/demo/foo",
			},
			EncodeFlags: EncodeDomain | EncodePath,
			Expected:    "gno.land/r/demo/foo",
		},

		{
			Name: "Encode Path Only",
			GnoURL: GnoURL{
				Path: "/r/demo/foo",
			},
			EncodeFlags: EncodePath,
			Expected:    "/r/demo/foo",
		},

		{
			Name: "Encode Path and File",
			GnoURL: GnoURL{
				Path: "/r/demo/foo",
				File: "render.gno",
			},
			EncodeFlags: EncodePath,
			Expected:    "/r/demo/foo/render.gno",
		},

		{
			Name: "Encode Path, File, and Args",
			GnoURL: GnoURL{
				Path: "/r/demo/foo",
				File: "render.gno",
				Args: "example",
			},
			EncodeFlags: EncodePath | EncodeArgs,
			Expected:    "/r/demo/foo/render.gno:example",
		},

		{
			Name: "Encode Path and Args",
			GnoURL: GnoURL{
				Path: "/r/demo/foo",
				Args: "example",
			},
			EncodeFlags: EncodePath | EncodeArgs,
			Expected:    "/r/demo/foo:example",
		},

		{
			Name: "Encode Path, Args, and WebQuery",
			GnoURL: GnoURL{
				Path: "/r/demo/foo",
				Args: "example",
				WebQuery: url.Values{
					"tz": []string{"Europe/Paris"},
				},
			},
			EncodeFlags: EncodePath | EncodeArgs | EncodeWebQuery,
			Expected:    "/r/demo/foo:example$tz=Europe%2FParis",
		},

		{
			Name: "Encode Full URL",
			GnoURL: GnoURL{
				Path: "/r/demo/foo",
				Args: "example",
				WebQuery: url.Values{
					"tz": []string{"Europe/Paris"},
				},
				Query: url.Values{
					"hello": []string{"42"},
				},
			},
			EncodeFlags: EncodePath | EncodeArgs | EncodeWebQuery | EncodeQuery,
			Expected:    "/r/demo/foo:example$tz=Europe%2FParis?hello=42",
		},

		{
			Name: "Encode Args and Query",
			GnoURL: GnoURL{
				Path: "/r/demo/foo",
				Args: "hello Jo$ny",
				Query: url.Values{
					"hello": []string{"42"},
				},
			},
			EncodeFlags: EncodeArgs | EncodeQuery,
			Expected:    "hello%20Jo%24ny?hello=42",
		},

		{
			Name: "Encode Args and Query (No Escape)",
			GnoURL: GnoURL{
				Path: "/r/demo/foo",
				Args: "hello Jo$ny",
				Query: url.Values{
					"hello": []string{"42"},
				},
			},
			EncodeFlags: EncodeArgs | EncodeQuery | EncodeNoEscape,
			Expected:    "hello Jo$ny?hello=42",
		},

		{
			Name: "Encode Args and Query",
			GnoURL: GnoURL{
				Path: "/r/demo/foo",
				Args: "example",
				Query: url.Values{
					"hello": []string{"42"},
				},
			},
			EncodeFlags: EncodeArgs | EncodeQuery,
			Expected:    "example?hello=42",
		},

		{
			Name: "Encode with Escaped Characters",
			GnoURL: GnoURL{
				Path: "/r/demo/foo",
				Args: "example with spaces",
				WebQuery: url.Values{
					"tz": []string{"Europe/Paris"},
				},
				Query: url.Values{
					"hello": []string{"42"},
				},
			},
			EncodeFlags: EncodePath | EncodeArgs | EncodeWebQuery | EncodeQuery,
			Expected:    "/r/demo/foo:example%20with%20spaces$tz=Europe%2FParis?hello=42",
		},

		{
			Name: "Encode Path, Args, and Query",
			GnoURL: GnoURL{
				Path: "/r/demo/foo",
				Args: "example",
				Query: url.Values{
					"hello": []string{"42"},
				},
			},
			EncodeFlags: EncodePath | EncodeArgs | EncodeQuery,
			Expected:    "/r/demo/foo:example?hello=42",
		},

		{
			Name: "WebQuery with empty value",
			GnoURL: GnoURL{
				Path: "/r/demo/foo",
				WebQuery: url.Values{
					"source": {""},
				},
			},
			EncodeFlags: EncodePath | EncodeWebQuery | EncodeNoEscape,
			Expected:    "/r/demo/foo$source",
		},

		{
			Name: "WebQuery with nil",
			GnoURL: GnoURL{
				Path: "/r/demo/foo",
				WebQuery: url.Values{
					"debug": nil,
				},
			},
			EncodeFlags: EncodePath | EncodeWebQuery,
			Expected:    "/r/demo/foo$",
		},

		{
			Name: "WebQuery with regular value",
			GnoURL: GnoURL{
				Path: "/r/demo/foo",
				WebQuery: url.Values{
					"key": {"value"},
				},
			},
			EncodeFlags: EncodePath | EncodeWebQuery,
			Expected:    "/r/demo/foo$key=value",
		},

		{
			Name: "WebQuery mixing empty and nil and filled values",
			GnoURL: GnoURL{
				Path: "/r/demo/foo",
				WebQuery: url.Values{
					"source": {""},
					"debug":  nil,
					"user":   {"Alice"},
				},
			},
			EncodeFlags: EncodePath | EncodeWebQuery,
			Expected:    "/r/demo/foo$source&user=Alice",
		},

		{
			Name: "WebQuery mixing nil and filled values",
			GnoURL: GnoURL{
				Path: "/r/demo/foo",
				WebQuery: url.Values{
					"debug": nil,
					"user":  {"Alice"},
				},
			},
			EncodeFlags: EncodePath | EncodeWebQuery,
			Expected:    "/r/demo/foo$user=Alice",
		},

		{
			Name: "Slashes in args are encoded by default",
			GnoURL: GnoURL{
				Path: "/r/demo/foo",
				Args: "path/to/resource",
			},
			EncodeFlags: EncodePath | EncodeArgs,
			Expected:    "/r/demo/foo:path%2Fto%2Fresource",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			result := tc.GnoURL.Encode(tc.EncodeFlags)
			require.True(t, tc.GnoURL.IsValidPath(), "gno url is not valid")
			assert.Equal(t, tc.Expected, result)
		})
	}
}

func TestGnoURL_Helpers(t *testing.T) {
	gurl, err := Parse("/r/demo/users:p/foo%3d%24bar%26%3fbaz$func=foo?hey=1&a%2bb=b%2b&%26=%3d")
	require.NoError(t, err)
	assert.Equal(t, "p/foo=$bar&?baz", gurl.Args)
	assert.Equal(t, url.Values{
		"hey": {"1"},
		"a+b": {"b+"},
		"&":   {"="},
	}, gurl.Query)

	// EncodeArgs is used to pass the values to the render function in Gno.
	// We are less restrictive about URLs in Gno, so it is only really important
	// that ? in the path is encoded correctly, but we can pass in $ and other symbols
	// regularly.
	// The queries should always be encoded.
	assert.Equal(t, "p/foo=$bar&%3Fbaz?%26=%3D&a%2Bb=b%2B&hey=1", gurl.EncodeArgs())
	// EncodeURL, instead, should behave as the above but include the path.
	// This should be used for display about the path to be used, but not used
	// as an actual URL in <a> links.
	assert.Equal(t, "/r/demo/users:p/foo=$bar&%3Fbaz?%26=%3D&a%2Bb=b%2B&hey=1", gurl.EncodeURL())
	// EncodeWebURL is the representation meant to be used in links.
	// WebQueries are included, and as such any webquery in the path should be escaped.
	// The path is escaped too, except for / which is converted back for ease of reading.
	assert.Equal(t, "/r/demo/users:p/foo=%24bar&%3Fbaz$func=foo?%26=%3D&a%2Bb=b%2B&hey=1", gurl.EncodeWebURL())
}

func TestEncodeFormURL(t *testing.T) {
	testCases := []struct {
		name     string
		gnoURL   GnoURL
		expected string
	}{
		{
			name: "simple path",
			gnoURL: GnoURL{
				Path: "/r/demo/foo",
				Args: "submit",
			},
			expected: "/r/demo/foo:submit",
		},
		{
			name: "path with slashes encoded",
			gnoURL: GnoURL{
				Path: "/r/demo/foo",
				Args: "path/to/resource",
			},
			expected: "/r/demo/foo:path%2Fto%2Fresource",
		},
		{
			name: "path traversal encoded",
			gnoURL: GnoURL{
				Path: "/r/demo/foo",
				Args: "../../../bar",
			},
			expected: "/r/demo/foo:..%2F..%2F..%2Fbar",
		},
		{
			name: "with query params",
			gnoURL: GnoURL{
				Path: "/r/demo/foo",
				Args: "foo/bar",
				Query: url.Values{
					"name": {"test"},
				},
			},
			expected: "/r/demo/foo:foo%2Fbar?name=test",
		},
		{
			name: "no args",
			gnoURL: GnoURL{
				Path: "/r/demo/foo",
				Query: url.Values{
					"name": {"test"},
				},
			},
			expected: "/r/demo/foo?name=test",
		},
		{
			name: "query in args is encoded",
			gnoURL: GnoURL{
				Path: "/r/demo/foo",
				Args: "submit?evil=injection",
			},
			// ? is encoded so browser won't parse as query string
			expected: "/r/demo/foo:submit%3Fevil=injection",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.gnoURL.EncodeFormURL()
			assert.Equal(t, tc.expected, result)
		})
	}
}
