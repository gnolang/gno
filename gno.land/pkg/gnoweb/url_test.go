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
			Err:      ErrURLMalformedPath,
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
			Input: "https://gno.land/r/demo/AAA$BBB:CCC&DDD$EEE",
			Err:   ErrURLMalformedPath, // `/r/demo/AAA$BBB` is an invalid path
		},

		{
			Name:  "args-webquery-args",
			Input: "https://gno.land/r/demo/AAA:BBB$CCC&DDD:EEE",
			Expected: &GnoURL{
				Domain: "gno.land",
				Path:   "/r/demo/AAA",
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
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Logf("testing input: %q", tc.Input)

			u, err := url.Parse(tc.Input)
			require.NoError(t, err)

			result, err := ParseGnoURL(u)
			if tc.Err == nil {
				require.NoError(t, err)
				t.Logf("encoded web path: %q", result.EncodeWebPath())
			} else {
				require.Error(t, err)
				require.ErrorIs(t, err, tc.Err)
			}

			assert.Equal(t, tc.Expected, result)
		})
	}
}
