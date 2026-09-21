package main

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestSplitURL(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		in, base, args, query string
	}{
		{"/r/x/y", "/r/x/y", "", ""},
		// the trailing slash is kept: gnoweb renders one and lists the other
		{"/r/x/y/", "/r/x/y/", "", ""},
		{"/r/x/y$source", "/r/x/y", "", "source"},
		{"/r/x/y$source&file=a.gno", "/r/x/y", "", "source&file=a.gno"},
		{"/r/x/y:p/about", "/r/x/y", "p/about", ""},
		{"/r/x/y:p/about$source", "/r/x/y", "p/about", "source"},
		{"/p/x/y", "/p/x/y", "", ""},
	} {
		base, args, query := splitURL(tc.in)
		if base != tc.base || args != tc.args || query != tc.query {
			t.Errorf("splitURL(%q) = %q,%q,%q; want %q,%q,%q",
				tc.in, base, args, query, tc.base, tc.args, tc.query)
		}
	}
}

func TestCanonicalURL(t *testing.T) {
	t.Parallel()
	// gnoweb templates emit the web query in both orders; they are one page.
	for _, tc := range []struct{ in, want string }{
		{"/r/x/y", "/r/x/y"},
		// the listing view keeps its slash: it is a different page
		{"/r/x/y/", "/r/x/y/"},
		{"/r/x/y$source&file=a.gno", "/r/x/y$file=a.gno&source"},
		{"/r/x/y$file=a.gno&source", "/r/x/y$file=a.gno&source"},
		{"/r/x/y:p/about$source", "/r/x/y:p/about$source"},
	} {
		if got := canonicalURL(tc.in); got != tc.want {
			t.Errorf("canonicalURL(%q) = %q; want %q", tc.in, got, tc.want)
		}
	}
}

func TestURLToFile(t *testing.T) {
	t.Parallel()
	// No output path may contain $, : or & — they survive a URL but not every
	// static host.
	for _, tc := range []struct{ in, want string }{
		{"/r/gnoland/home", "r/gnoland/home/index.html"},
		{"/r/gnoland/home$source", "r/gnoland/home/_t/source/index.html"},
		// slugs that are not already a safe segment carry a digest of the input,
		// so ":p/a-b", ":p/a/b" and ":p/a&b" cannot share a file
		{"/r/gnoland/home$file=home.gno&source", "r/gnoland/home/_t/file-home.gno-source-2ca84ba4/index.html"},
		{"/r/gnoland/blog:p/hello", "r/gnoland/blog/_a/p-hello-13cc55eb/index.html"},
		{"/r/gnoland/blog:p/hello$source", "r/gnoland/blog/_a/p-hello-13cc55eb/_t/source/index.html"},
		{"/r/", "r/_dir/index.html"},
		{"/", "_root/index.html"},
	} {
		if got := urlToFile(tc.in); got != tc.want {
			t.Errorf("urlToFile(%q) = %q; want %q", tc.in, got, tc.want)
		}
	}
}

func TestCrawlerInScope(t *testing.T) {
	t.Parallel()
	// Budgeted so the file case still exercises the tab parsing; which files are
	// worth keeping is covered by TestWantFile and TestFileBudgetZeroDropsAllUnchanged.
	c := &Crawler{
		Realms:     []string{"gno.land/r/gnoland/home", "gno.land/r/demo/counter"},
		FileBudget: GnowebFileBudget,
	}
	for _, tc := range []struct {
		in   string
		want bool
	}{
		{"/r/gnoland/home", true},
		{"/r/gnoland/home$source", true},
		{"/r/gnoland/home$file=home.gno&source", true},
		// "/r/x/y/" is gnoweb's listing, a different page from the render.
		{"/r/gnoland/home/", false},
		{"/r/gnoland/home:p/x", true},
		{"/r/demo/counter$help", true},
		// out of the selected set
		{"/r/gnoland/blog", false},
		{"/r/", false},
		{"/u/g1abc", false},
		// bounded: these enumerate the object graph or are not pages at all
		{"/r/gnoland/home$state", false},
		{"/r/gnoland/home$state&oid=deadbeef%3A10", false},
		{"/r/gnoland/home$state&tid=errors.errorString", false},
		{"/r/gnoland/home$download&file=home.gno", false},
		{"/r/gnoland/home$help&func=Admin", false},
		// $source/$help do not vary with render args — the tab-less page covers it
		{"/r/gnoland/home:p/x$source", false},
	} {
		if got := c.inScope(tc.in); got != tc.want {
			t.Errorf("inScope(%q) = %v; want %v", tc.in, got, tc.want)
		}
	}
}

func TestLinks(t *testing.T) {
	t.Parallel()
	body := `<a href="/r/x/y$source&amp;file=a.gno">s</a>
	         <a href="/r/x/y#frag">f</a>
	         <a href="https://gno.land/out">ext</a>
	         <a href="#local">l</a>
	         <link href="/public/main.css?v=1">
	         <script src="/public/js/index.js"></script>`
	want := []string{"/r/x/y", "/r/x/y$file=a.gno&source"}
	if got := links(body); !reflect.DeepEqual(got, want) {
		t.Errorf("links() = %v; want %v", got, want)
	}
}

func TestMapURL(t *testing.T) {
	t.Parallel()
	c := &Crawler{Live: "https://gno.land", pages: map[string]*page{
		"/r/x/y":                   {File: "r/x/y/index.html"},
		"/r/x/y$source":            {File: "r/x/y/_t/source/index.html"},
		"/r/x/y$file=a.gno&source": {File: "r/x/y/_t/file-a.gno-source/index.html"},
	}}
	const up = "../../../" // a page at r/x/y/index.html
	for _, tc := range []struct{ in, want string }{
		// captured pages become relative
		{"/r/x/y", "../../../r/x/y/"},
		{"/r/x/y$source", "../../../r/x/y/_t/source/"},
		{"/r/x/y$source&file=a.gno", "../../../r/x/y/_t/file-a.gno-source/"},
		{"/r/x/y$source#L14", "../../../r/x/y/_t/source/#L14"},
		// assets keep their cache-buster and point at the copied tree
		{"/public/main.css?v=1", "../../../public/main.css?v=1"},
		{"/public//favicon.ico", "../../../public/favicon.ico"},
		// the site root is the snapshot index
		{"/", "../../../index.html"},
		// anything not captured goes to the live site
		{"/r/other/realm", "https://gno.land/r/other/realm"},
		{"/u/g1abc", "https://gno.land/u/g1abc"},
		// untouched
		{"#anchor", "#anchor"},
		{"https://example.com", "https://example.com"},
		{"", ""},
	} {
		if got := c.mapURL(tc.in, up); got != tc.want {
			t.Errorf("mapURL(%q) = %q; want %q", tc.in, got, tc.want)
		}
	}
}

func TestURLOf(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct{ in, want string }{
		{"gno.land/r/gnoland/home", "/r/gnoland/home"},
		{"gno.land/p/nt/avl/v0", "/p/nt/avl/v0"},
	} {
		if got := urlOf(tc.in); got != tc.want {
			t.Errorf("urlOf(%q) = %q; want %q", tc.in, got, tc.want)
		}
	}
}

func TestRewriteAddsNoindex(t *testing.T) {
	t.Parallel()
	c := &Crawler{Live: "https://gno.land", pages: map[string]*page{}}
	for _, tc := range []struct{ name, body string }{
		{"normal head", `<!doctype html><html><head><title>x</title></head><body>hi</body></html>`},
		{"head with attrs", `<html><head lang="en"><title>x</title></head></html>`},
		// A page gnoweb serves without a <head> (an error view, say) must still
		// carry the tag rather than silently become indexable.
		{"no head", `<p>fragment</p>`},
		// gnoweb's own layout ships this on every page; it must be replaced,
		// not joined by a second, contradicting tag.
		{"gnoweb's index,follow", `<html><head><meta name="robots" content="index, follow" /><title>x</title></head></html>`},
	} {
		got := c.rewrite(&page{File: "r/x/index.html", Body: tc.body})
		if !strings.Contains(got, noindexTag) {
			t.Errorf("%s: rewrite dropped the noindex tag: %s", tc.name, got)
		}
		if n := strings.Count(got, noindexTag); n != 1 {
			t.Errorf("%s: noindex tag appears %d times, want 1", tc.name, n)
		}
		if n := strings.Count(strings.ToLower(got), `<meta name="robots"`); n != 1 {
			t.Errorf("%s: %d robots metas, want exactly 1 — conflicting directives", tc.name, n)
		}
		if strings.Contains(got, "index, follow") {
			t.Errorf("%s: gnoweb's index,follow survived: %s", tc.name, got)
		}
	}
}

func TestWantFile(t *testing.T) {
	t.Parallel()
	c := &Crawler{
		Realms:       []string{"gno.land/r/x/touched", "gno.land/r/x/untouched"},
		ChangedFiles: map[string][]string{"gno.land/r/x/touched": {"a.gno", "gnomod.toml"}},
		FileBudget:   GnowebFileBudget,
	}
	// A realm the PR edited: exactly its changed files, nothing else.
	for _, tc := range []struct {
		url  string
		want bool
	}{
		{"/r/x/touched$source&file=a.gno", true},
		{"/r/x/touched$source&file=gnomod.toml", true},
		{"/r/x/touched$source&file=b.gno", false},
		{"/r/x/touched$source", true}, // the overview always stays
	} {
		if got := c.inScope(tc.url); got != tc.want {
			t.Errorf("inScope(%q) = %v; want %v", tc.url, got, tc.want)
		}
	}
	// A realm pulled in by a dependency: a small budget, then nothing, so one
	// widely imported package cannot drag in a page per file per realm.
	for i := range GnowebFileBudget {
		u := fmt.Sprintf("/r/x/untouched$source&file=f%d.gno", i)
		if !c.inScope(u) || !c.charge(u) {
			t.Errorf("%q refused within the budget", u)
		}
	}
	if c.charge("/r/x/untouched$source&file=over.gno") {
		t.Error("budget exceeded but the page was still captured")
	}
}

func TestFileBudgetZeroDropsAllUnchanged(t *testing.T) {
	t.Parallel()
	// A dependency bump: no file in the realm changed, and no gnoweb change to
	// justify looking at one. Per-file pages are 60% of a wide preview's bytes.
	c := &Crawler{Realms: []string{"gno.land/r/x/dep"}}
	if c.inScope("/r/x/dep$source&file=any.gno") {
		t.Error("per-file page captured with a zero budget")
	}
	if !c.inScope("/r/x/dep$source") {
		t.Error("the source overview must still be captured")
	}
}

func TestURLToFileIsSafe(t *testing.T) {
	t.Parallel()
	// gnoweb serves "/r/x/y:..", and path.Join would fold ".." onto the realm's
	// own render page. No output may escape its realm directory.
	for _, u := range []string{"/r/x/y:..", "/r/x/y$..", "/r/x/y:../..", "/r/x/y:."} {
		got := urlToFile(u)
		if got == "r/x/y/index.html" {
			t.Errorf("urlToFile(%q) = %q — traversed onto the render page", u, got)
		}
		if !strings.HasPrefix(got, "r/x/y/") {
			t.Errorf("urlToFile(%q) = %q — escaped the realm directory", u, got)
		}
		if strings.Contains(got, "/../") || strings.HasSuffix(got, "/..") {
			t.Errorf("urlToFile(%q) = %q — contains a parent reference", u, got)
		}
	}

	// Distinct URLs must not share a file.
	seen := map[string]string{}
	for _, u := range []string{
		"/r/x/y:p/a-b", "/r/x/y:p/a/b", "/r/x/y:p/a&b",
		"/r/x/y", "/r/x/y/", "/r/x/y$source",
	} {
		f := urlToFile(u)
		if prev, dup := seen[f]; dup {
			t.Errorf("urlToFile(%q) collides with %q on %q", u, prev, f)
		}
		seen[f] = u
	}

	// A long argument must not produce an unbounded path segment.
	long := urlToFile("/r/x/y:" + strings.Repeat("z", 400))
	for seg := range strings.SplitSeq(long, "/") {
		if len(seg) > maxSlugLen+16 {
			t.Errorf("segment %q is %d chars", seg, len(seg))
		}
	}
}

func TestSplitURLKeepsTrailingSlash(t *testing.T) {
	t.Parallel()
	// Measured on gnoweb: /r/gnoland/home is 62,174 bytes and /r/gnoland/home/
	// is 48,642 — different pages. Trimming the slash merged them onto one file.
	a, _, _ := splitURL("/r/x/y")
	b, _, _ := splitURL("/r/x/y/")
	if a == b {
		t.Fatalf("splitURL collapsed the listing onto the render: both %q", a)
	}
}

func TestSeedsSkipSingleSegmentDirs(t *testing.T) {
	t.Parallel()
	// gnoweb answers "/r" with 400; seeding it only buys a logged error.
	c := &Crawler{Realms: []string{"gno.land/r/gnoland/home"}}
	for _, s := range c.Seeds() {
		if s == "/r" || s == "/p" {
			t.Errorf("Seeds() includes %q, which gnoweb answers with 400", s)
		}
	}
}

func TestFileBudgetChargedOnCaptureNotDiscovery(t *testing.T) {
	t.Parallel()
	c := &Crawler{Realms: []string{"gno.land/r/x/y"}, FileBudget: 1}
	// Discovery alone must not spend the budget: a link that 404s would
	// otherwise consume the slot a real page needed.
	for range 5 {
		if !c.inScope("/r/x/y$source&file=a.gno") {
			t.Fatal("discovery was refused before anything was captured")
		}
	}
	if !c.charge("/r/x/y$source&file=a.gno") {
		t.Fatal("first capture refused")
	}
	if c.charge("/r/x/y$source&file=b.gno") {
		t.Error("second capture accepted with a budget of 1")
	}
}

// Every other axis of the crawl is bounded by the realm: it has the files it
// has, and gnoweb defines the tabs. Render arguments are not: a realm may link
// one page per value it knows about, and each one is a full page of chrome.
// Nothing in examples/ does that today (the busiest renders 5), but a realm
// that starts must not be able to exhaust -max-pages, which fails the whole
// render rather than trimming it.
func TestRenderArgumentsAreCappedPerRealm(t *testing.T) {
	c := &Crawler{Realms: []string{"gno.land/r/demo/counter"}, ArgBudget: 2}
	for i, u := range []string{"/r/demo/counter:1", "/r/demo/counter:2"} {
		if !c.charge(u) {
			t.Fatalf("argument page %d dropped while the budget had room", i)
		}
	}
	if c.charge("/r/demo/counter:3") {
		t.Fatal("the argument budget did not stop the third page")
	}
	// The budget is per realm, not global.
	other := &Crawler{Realms: []string{"gno.land/r/demo/counter", "gno.land/r/demo/other"}, ArgBudget: 1}
	if !other.charge("/r/demo/counter:1") || !other.charge("/r/demo/other:1") {
		t.Fatal("one realm's arguments were charged to another")
	}
	// A realm's own render page and its tabs are never argument pages, and a
	// zero budget is no cap at all (as for MaxPages and MaxRealms), so a
	// Crawler nobody configured behaves exactly as it did before this budget.
	free := &Crawler{Realms: []string{"gno.land/r/demo/counter"}, ArgBudget: 0}
	if !free.charge("/r/demo/counter:1") || !free.charge("/r/demo/counter:2") {
		t.Fatal("a zero ArgBudget dropped argument pages instead of meaning no cap")
	}
	for _, u := range []string{"/r/demo/counter", "/r/demo/counter$source", "/r/demo/counter$help"} {
		if !free.charge(u) {
			t.Fatalf("%s was charged as an argument page", u)
		}
	}
	// And a directory page belongs to no realm, so it is never charged.
	if !free.charge("/r/demo") {
		t.Fatal("a directory page was charged to a realm")
	}
}

// inScope must stop following argument links once the budget is spent, so the
// queue does not grow with pages charge() will only throw away.
func TestInScopeStopsFollowingArgumentsOnceSpent(t *testing.T) {
	c := &Crawler{Realms: []string{"gno.land/r/demo/counter"}, ArgBudget: 1}
	if !c.inScope("/r/demo/counter:1") {
		t.Fatal("the first argument page was refused")
	}
	c.charge("/r/demo/counter:1")
	if c.inScope("/r/demo/counter:2") {
		t.Fatal("a second argument page was still followed after the budget was spent")
	}
	// The argument-free views stay in scope regardless.
	for _, u := range []string{"/r/demo/counter", "/r/demo/counter$source"} {
		if !c.inScope(u) {
			t.Fatalf("%s left scope with the argument budget spent", u)
		}
	}
}
