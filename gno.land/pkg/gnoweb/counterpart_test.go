package gnoweb

import (
	"errors"
	"strconv"
	"testing"
	"time"

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
			name:   "twin without siblings opens directly",
			twin:   "/p/alice/golf/game",
			paths:  []string{"/p/alice/golf/game", "/p/alice/golf/ui/board"},
			target: "/p/alice/golf/game", n: 1,
		},
		{
			name:   "twin with siblings opens their listing",
			twin:   "/p/alice/golf/v0",
			paths:  []string{"/p/alice/golf/v0", "/p/alice/golf/v2"},
			target: "/p/alice/golf", n: 2,
		},
		{
			name:   "twin's listing counts the whole subtree it shows",
			twin:   "/p/alice/golf/v0",
			paths:  []string{"/p/alice/golf/v0", "/p/alice/golf/v2", "/p/alice/golf/ui/board"},
			target: "/p/alice/golf", n: 3,
		},
		{
			name:   "twin's directory being a package opens its Directories section",
			twin:   "/p/alice/golf/v1",
			paths:  []string{"/p/alice/golf", "/p/alice/golf/v1", "/p/alice/golf/v2", "/p/alice/golf/v2/x"},
			target: "/p/alice/golf$source#subpackages", n: 2,
		},
		{
			name:   "twin at the project root opens directly",
			twin:   "/p/alice/golf",
			paths:  []string{"/p/alice/golf", "/p/alice/golf/ui"},
			target: "/p/alice/golf", n: 1,
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
			name:   "a package above the twin is named alone",
			twin:   "/p/alice/golf/utils",
			paths:  []string{"/p/alice/golf", "/p/alice/golf/impl/v0", "/p/alice/golf/init/v0"},
			target: "/p/alice/golf", n: 1,
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
	assert.Equal(t, "Matching package", one.Label)
	assert.Equal(t, "ico-pure", one.Icon)
	assert.Equal(t, "/p/alice/golf/physics", one.URL)

	many := counterpartLink("/r/alice/golf", "/r/alice/golf", 3)
	assert.Equal(t, "3 matching realms", many.Label)
	assert.Equal(t, "ico-realm", many.Icon)
	assert.Equal(t, "/r/alice/golf", many.URL)

	capped := counterpartLink("/p/alice/golf", "/p/alice/golf", maxCounterpartPaths)
	assert.Equal(t, "100+ matching packages", capped.Label)
}

func TestCounterpartCache(t *testing.T) {
	t.Parallel()

	now := time.Unix(0, 0)
	c := counterpartCache{now: func() time.Time { return now }}
	calls := 0
	list := func() ([]string, error) {
		calls++
		return nil, nil // an empty answer is cached too
	}

	c.get("/p/alice/golf", list)
	c.get("/p/alice/golf", list)
	assert.Equal(t, 1, calls)

	now = now.Add(counterpartTTL)
	c.get("/p/alice/golf", list)
	assert.Equal(t, 2, calls)

	_, err := c.get("/p/bob/x", func() ([]string, error) { return nil, errors.New("node down") })
	assert.Error(t, err)
	c.get("/p/bob/x", list)
	assert.Equal(t, 3, calls, "errors are not cached")
}

func TestCounterpartCacheFull(t *testing.T) {
	t.Parallel()

	now := time.Unix(0, 0)
	c := counterpartCache{now: func() time.Time { return now }}
	for i := range maxCounterpartEntries {
		c.store(strconv.Itoa(i), nil)
	}

	c.store("late", nil)
	_, ok := c.lookup("late")
	assert.False(t, ok, "a full cache of live entries keeps no more")

	now = now.Add(counterpartTTL)
	c.store("late", nil)
	_, ok = c.lookup("late")
	assert.True(t, ok, "expired entries make room")
	assert.Len(t, c.entries, 1)
}
