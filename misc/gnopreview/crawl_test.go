package main

import (
	"reflect"
	"testing"
)

func TestSplitURL(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		in, base, args, query string
	}{
		{"/r/x/y", "/r/x/y", "", ""},
		{"/r/x/y/", "/r/x/y", "", ""},
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
		{"/r/x/y/", "/r/x/y"},
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
		{"/r/gnoland/home$file=home.gno&source", "r/gnoland/home/_t/file-home.gno-source/index.html"},
		{"/r/gnoland/blog:p/hello", "r/gnoland/blog/_a/p-hello/index.html"},
		{"/r/gnoland/blog:p/hello$source", "r/gnoland/blog/_a/p-hello/_t/source/index.html"},
		{"/r/", "r/index.html"},
		{"/", "_root/index.html"},
	} {
		if got := urlToFile(tc.in); got != tc.want {
			t.Errorf("urlToFile(%q) = %q; want %q", tc.in, got, tc.want)
		}
	}
}

func TestCrawlerInScope(t *testing.T) {
	t.Parallel()
	c := &Crawler{Realms: []string{"gno.land/r/gnoland/home", "gno.land/r/demo/counter"}}
	for _, tc := range []struct {
		in   string
		want bool
	}{
		{"/r/gnoland/home", true},
		{"/r/gnoland/home$source", true},
		{"/r/gnoland/home$file=home.gno&source", true},
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
