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
		Err      *error
	}{
		{
			Name:     "malformed url",
			Input:    "https://gno.land/r/demo/:$?",
			Expected: nil,
			Err:      &ErrURLMalformedPath,
		},
		{
			Name:     "malformed url 2",
			Input:    "https://gno.land/r/dem)o:$?",
			Expected: nil,
			Err:      &ErrURLMalformedPath,
		},

		{
			Name:  "webquery + query",
			Input: "https://gno.land/r/demo/foo$help&func=Bar&name=Baz",
			Expected: &GnoURL{
				Path:     "/r/demo/foo",
				Kind:     KindRealm,
				PathArgs: "",
				WebQuery: url.Values{
					"help": []string{""},
					"func": []string{"Bar"},
					"name": []string{"Baz"},
				},
				Query: url.Values{},
			},
			Err: nil,
		},

		{
			Name:  "path args + webquery",
			Input: "https://gno.land/r/demo/foo:example$tz=Europe/Paris",
			Expected: &GnoURL{
				Path:     "/r/demo/foo",
				Kind:     KindRealm,
				PathArgs: "example",
				WebQuery: url.Values{
					"tz": []string{"Europe/Paris"},
				},
				Query: url.Values{},
			},
			Err: nil,
		},

		{
			Name:  "path args + webquery + query",
			Input: "https://gno.land/r/demo/foo:example$tz=Europe/Paris?hello=42",
			Expected: &GnoURL{
				Path:     "/r/demo/foo",
				Kind:     KindRealm,
				PathArgs: "example",
				WebQuery: url.Values{
					"tz": []string{"Europe/Paris"},
				},
				Query: url.Values{
					"hello": []string{"42"},
				},
			},
			Err: nil,
		},

		{
			Name:  "webquery inside query",
			Input: "https://gno.land/r/demo/foo:example?value=42$tz=Europe/Paris",
			Expected: &GnoURL{
				Path:     "/r/demo/foo",
				Kind:     KindRealm,
				PathArgs: "example",
				WebQuery: url.Values{},
				Query: url.Values{
					"value": []string{"42$tz=Europe/Paris"},
				},
			},
			Err: nil,
		},

		{
			Name:  "webquery escaped $",
			Input: "https://gno.land/r/demo/foo:example%24hello=43$hello=42",
			Expected: &GnoURL{
				Path:     "/r/demo/foo",
				Kind:     KindRealm,
				PathArgs: "example$hello=43",
				WebQuery: url.Values{
					"hello": []string{"42"},
				},
				Query: url.Values{},
			},
			Err: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			u, err := url.Parse(tc.Input)
			require.NoError(t, err)

			result, err := ParseGnoURL(u)
			if tc.Err == nil {
				require.NoError(t, err)
				t.Logf("parsed: %s", result.EncodePath())
				t.Logf("parsed web: %s", result.EncodeWebPath())
			} else {
				require.ErrorAs(t, err, tc.Err)
			}

			assert.Equal(t, tc.Expected, result)
		})

	}
}
