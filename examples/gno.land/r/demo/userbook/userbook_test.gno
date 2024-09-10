package userbook

import (
	"std"
	"strings"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/ufmt"
)

func TestRender(t *testing.T) {
	// Sign up 20 users + deployer
	for i := 0; i < 20; i++ {
		addrName := ufmt.Sprintf("test%d", i)
		caller := testutils.TestAddress(addrName)
		std.TestSetOrigCaller(caller)
		SignUp()
	}

	testCases := []struct {
		name                  string
		nextPage              bool
		prevPage              bool
		path                  string
		expectedNumberOfUsers int
	}{
		{
			name:                  "1st page render",
			nextPage:              true,
			prevPage:              false,
			path:                  "page/1",
			expectedNumberOfUsers: 20,
		},
		{
			name:                  "2nd page render",
			nextPage:              false,
			prevPage:              true,
			path:                  "page/2",
			expectedNumberOfUsers: 1,
		},
		{
			name:                  "Invalid path render",
			nextPage:              true,
			prevPage:              false,
			path:                  "page/invalidtext",
			expectedNumberOfUsers: 20,
		},
		{
			name:                  "Empty Page",
			nextPage:              false,
			prevPage:              false,
			path:                  "page/1000",
			expectedNumberOfUsers: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := Render(tc.path)
			numUsers := countUsers(got)

			if tc.prevPage && !strings.Contains(got, "Previous page") {
				t.Fatalf("expected to find Previous page, didn't find it")
			}
			if tc.nextPage && !strings.Contains(got, "Next page") {
				t.Fatalf("expected to find Next page, didn't find it")
			}

			if tc.expectedNumberOfUsers != numUsers {
				t.Fatalf("expected %d, got %d users", tc.expectedNumberOfUsers, numUsers)
			}
		})
	}
}

func countUsers(input string) int {
	return strings.Count(input, "#### User #")
}
