package gnoweb

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseGnoURL(t *testing.T) {
	testCases := []struct {
		Name     string
		Input    string
		Expected *GnoURL
		Err      error
	}{
		{
			Name:     "malformed url",
			Input:    "https://gno.land/r/dem)o:$?",
			Expected: nil,
			Err:      ErrURLInvalidPath,
		},

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
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Logf("testing input: %q", tc.Input)

			u, err := url.Parse(tc.Input)
			require.NoError(t, err)

			result, err := ParseGnoURL(u)
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

func TestEncode(t *testing.T) {
	testCases := []struct {
		Name        string
		GnoURL      GnoURL
		EncodeFlags EncodeFlag
		Expected    string
	}{
		{
			Name: "encode domain",
			GnoURL: GnoURL{
				Domain: "gno.land",
				Path:   "/r/demo/foo",
			},
			EncodeFlags: EncodeDomain,
			Expected:    "gno.land",
		},

		{
			Name: "encode web query without escape",
			GnoURL: GnoURL{
				Domain: "gno.land",
				Path:   "/r/demo/foo",
				WebQuery: url.Values{
					"help":  []string{""},
					"fun$c": []string{"B$ ar"},
				},
			},
			EncodeFlags: EncodeWebQuery | EncodeNoEscape,
			Expected:    "$fun$c=B$ ar&help=",
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
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			result := tc.GnoURL.Encode(tc.EncodeFlags)
			require.True(t, tc.GnoURL.IsValid(), "gno url is not valid")
			assert.Equal(t, tc.Expected, result)
		})
	}
}
