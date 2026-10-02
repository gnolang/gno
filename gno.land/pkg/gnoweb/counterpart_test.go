package gnoweb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCounterpartRoots(t *testing.T) {
	t.Parallel()

	cases := []struct {
		path       string
		twin, root string
		ok         bool
	}{
		{"/r/alice/golf/game", "/p/alice/golf/game", "/p/alice/golf", true},
		{"/p/alice/golf/physics", "/r/alice/golf/physics", "/r/alice/golf", true},
		{"/r/alice/golf", "/p/alice/golf", "/p/alice/golf", true},
		{"/r/alice/golf/", "/p/alice/golf", "/p/alice/golf", true}, // directory URL
		{"/r/alice", "", "", false},                                // a namespace is /u/alice's job
		{"/r/", "", "", false},
		{"/u/alice/golf", "", "", false},
		{"/r/alice//golf", "", "", false},
		{"/", "", "", false},
	}
	for _, tc := range cases {
		twin, root, ok := counterpartRoots(tc.path)
		assert.Equal(t, tc.ok, ok, tc.path)
		assert.Equal(t, tc.twin, twin, tc.path)
		assert.Equal(t, tc.root, root, tc.path)
	}
}

func TestCounterpartTarget(t *testing.T) {
	t.Parallel()

	const root = "/p/alice/golf"
	cases := []struct {
		name   string
		twin   string
		paths  []string
		target string
		n      int
	}{
		{
			name:   "twin exists",
			twin:   "/p/alice/golf/game",
			paths:  []string{"/p/alice/golf/course", "/p/alice/golf/game"},
			target: "/p/alice/golf/game", n: 1,
		},
		{
			name:   "one package opens directly",
			twin:   "/p/alice/golf/game",
			paths:  []string{"/p/alice/golf/physics"},
			target: "/p/alice/golf/physics", n: 1,
		},
		{
			name:   "several packages open the project listing",
			twin:   "/p/alice/golf/game",
			paths:  []string{"/p/alice/golf/course", "/p/alice/golf/physics"},
			target: "/p/alice/golf", n: 2,
		},
		{
			name: "deepest directory holding packages wins",
			twin: "/p/alice/golf/engine/game",
			paths: []string{
				"/p/alice/golf/engine/ball",
				"/p/alice/golf/engine/club",
				"/p/alice/golf/ui",
			},
			target: "/p/alice/golf/engine", n: 2,
		},
		{
			name:  "a sibling sharing the prefix is not the project",
			twin:  "/p/alice/golf/game",
			paths: []string{"/p/alice/golfer", "/p/alice/golf2/x"},
		},
		{
			name: "malformed paths from the node are dropped",
			twin: "/p/alice/golf/game",
			paths: []string{
				"/p/alice/golf/../../../evil",
				`/p/alice/golf/"onmouseover=x`,
				"/p/alice/golf//x",
				"/p/alice/golf/Upper",
			},
		},
		{
			name:  "empty listing",
			twin:  "/p/alice/golf/game",
			paths: []string{""},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			target, n := counterpartTarget(tc.twin, root, tc.paths)
			assert.Equal(t, tc.target, target)
			assert.Equal(t, tc.n, n)
		})
	}
}

func TestCounterpartLink(t *testing.T) {
	t.Parallel()

	one := counterpartLink("/p/alice/golf/physics", "/p/alice/golf", 1)
	assert.Equal(t, "Package", one.Label)
	assert.Equal(t, "ico-pure", one.Icon)
	assert.Equal(t, "/p/alice/golf/physics", one.Title)

	many := counterpartLink("/r/alice/golf", "/r/alice/golf", 3)
	assert.Equal(t, "Realms", many.Label)
	assert.Equal(t, "ico-realm", many.Icon)
	assert.Equal(t, "/r/alice/golf (3 packages)", many.Title)

	capped := counterpartLink("/p/alice/golf", "/p/alice/golf", maxCounterpartPaths)
	assert.Equal(t, "/p/alice/golf (100+ packages)", capped.Title)
}
