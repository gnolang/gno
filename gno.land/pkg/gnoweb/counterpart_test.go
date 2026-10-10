package gnoweb

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
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
			name:   "a package above the twin with matching children opens its Directories section",
			twin:   "/p/alice/golf/foo",
			paths:  []string{"/p/alice/golf", "/p/alice/golf/v1", "/p/alice/golf/v2"},
			target: "/p/alice/golf$source#subpackages", n: 2,
		},
		{
			name:   "a package above the twin with one matching child is named alone",
			twin:   "/p/alice/golf/foo",
			paths:  []string{"/p/alice/golf", "/p/alice/golf/v1", "/p/alice/golf/ui/board"},
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

// TestCounterpartTargetDeepPathAllocs pins the twinless walk on a request
// path near the 4096-byte limit: it climbs one level at a time and checks
// every member at each, so an allocation per check cost hundreds of MB per
// request. Filtering the members may allocate (regexp state, more of it under
// -race), but nothing may allocate per level. Not parallel: AllocsPerRun
// counts every allocation in the process.
func TestCounterpartTargetDeepPathAllocs(t *testing.T) {
	const levels = 2040
	twin, root, ok := counterpartRoots("/r/alice/golf" + strings.Repeat("/a", levels))
	if !assert.True(t, ok) {
		return
	}
	paths := make([]string, maxCounterpartPaths)
	for i := range paths {
		paths[i] = fmt.Sprintf("/p/alice/golf/m%03d", i)
	}
	allocs := testing.AllocsPerRun(10, func() { counterpartTarget(twin, root, paths) })
	assert.Less(t, allocs, float64(levels))
}

func TestCounterpartLink(t *testing.T) {
	t.Parallel()

	one := counterpartLink("/p/alice/golf/physics", "/p/alice/golf", 1, false)
	assert.Equal(t, "Matching package", one.Label)
	assert.Equal(t, "ico-pure", one.Icon)
	assert.Equal(t, "/p/alice/golf/physics", one.URL)

	many := counterpartLink("/r/alice/golf", "/r/alice/golf", 3, false)
	assert.Equal(t, "3 matching realms", many.Label)
	assert.Equal(t, "ico-realm", many.Icon)
	assert.Equal(t, "/r/alice/golf", many.URL)

	section := counterpartLink("/r/tests/vm$source#subpackages", "/r/tests/vm", 2, false)
	assert.Equal(t, "/r/tests/vm$source#subpackages", section.URL)
	assert.Equal(t, "/r/tests/vm", section.Path)

	capped := counterpartLink("/p/alice/golf", "/p/alice/golf", maxCounterpartPaths, true)
	assert.Equal(t, "100+ matching packages", capped.Label)

	// A cut listing makes any count a floor, not only one at the cap.
	cut := counterpartLink("/p/alice/golf$source#subpackages", "/p/alice/golf", 99, true)
	assert.Equal(t, "99+ matching packages", cut.Label)
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

// Lookups outlive their page while holding a shared RPC slot, so a burst of
// new roots against a slow node must not run more than maxCounterpartLookups
// at once. Reported by gfanton on #6262.
func TestCounterpartCacheCapsLookupsInFlight(t *testing.T) {
	t.Parallel()

	var c counterpartCache
	entered := make(chan struct{})
	release := make(chan struct{})
	slow := func() ([]string, error) {
		entered <- struct{}{}
		<-release
		return nil, nil
	}

	var wg sync.WaitGroup
	for i := range maxCounterpartLookups {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.get("/p/slow/"+strconv.Itoa(i), slow)
		}()
		<-entered
	}

	listed := false
	_, err := c.get("/p/alice/golf", func() ([]string, error) {
		listed = true
		return nil, nil
	})
	assert.ErrorIs(t, err, errCounterpartBusy)
	assert.False(t, listed, "a lookup past the cap must not reach the node")

	close(release)
	wg.Wait()
	_, err = c.get("/p/alice/golf", func() ([]string, error) { return nil, nil })
	assert.NoError(t, err, "a freed slot serves the next view")
}
