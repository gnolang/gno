package gnoweb

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommunityIndexText(t *testing.T) {
	t.Parallel()

	for _, want := range []CommunityIndex{IndexNoCommunity, IndexRegisteredCommunity, IndexAllCommunity} {
		t.Run(want.String(), func(t *testing.T) {
			t.Parallel()

			text, err := want.MarshalText()
			require.NoError(t, err)
			var got CommunityIndex
			require.NoError(t, got.UnmarshalText(text))
			assert.Equal(t, want, got)
		})
	}

	t.Run("unknown value", func(t *testing.T) {
		t.Parallel()

		var got CommunityIndex
		assert.Error(t, got.UnmarshalText([]byte("some")))
	})
}

// TestFinalRobots checks what can lower a page's robots once it is served.
func TestFinalRobots(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		sp     servedPage
		header string // X-Robots-Tag a branch already set
		status int
		want   robots
	}{
		{"granted", servedPage{robots: indexFollow}, "", http.StatusOK, indexFollow},
		{"an error", servedPage{robots: indexFollow}, "", http.StatusNotFound, noIndexNoFollow},
		{"an empty page", servedPage{robots: indexFollow, render: pageRender{empty: true}}, "", http.StatusOK, noIndexNoFollow},
		{"the state explorer asked for less", servedPage{robots: noIndexFollow}, "noindex, nofollow", http.StatusOK, noIndexNoFollow},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rr := httptest.NewRecorder()
			if tc.header != "" {
				rr.Header().Set("X-Robots-Tag", tc.header)
			}
			assert.Equal(t, tc.want, finalRobots(rr, &tc.sp, tc.status))
			if tc.want != indexFollow {
				assert.Equal(t, tc.want.String(), rr.Header().Get("X-Robots-Tag"))
			}
		})
	}
}
