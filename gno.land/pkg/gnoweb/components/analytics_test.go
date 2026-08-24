package components

import (
	"net/url"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
)

func TestClassifyPageType(t *testing.T) {
	cases := []struct {
		name string
		mode ViewMode
		view ViewType
		want string
	}{
		{"source view wins over realm mode", ViewModeRealm, SourceViewType, "source"},
		{"help view wins over realm mode", ViewModeRealm, HelpViewType, "help"},
		{"directory view", ViewModeExplorer, DirectoryViewType, "directory"},
		{"status view", ViewModeExplorer, StatusViewType, "status"},
		{"redirect view", ViewModeExplorer, RedirectViewType, "redirect"},
		{"realm view falls through to mode", ViewModeRealm, RealmViewType, "realm"},
		{"home mode", ViewModeHome, RealmViewType, "home"},
		{"user mode", ViewModeUser, UserViewType, "user"},
		{"package mode", ViewModePackage, RealmViewType, "pure"},
		{"explorer mode", ViewModeExplorer, RealmViewType, "explorer"},
		{"unknown view + zero mode", ViewModeExplorer, ViewType("unknown"), "explorer"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ClassifyPageType(tc.mode, tc.view); got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestAnalyticsPath(t *testing.T) {
	cases := []struct {
		name string
		url  weburl.GnoURL
		want string
	}{
		{
			"plain realm route is kept",
			weburl.GnoURL{Path: "/r/demo/foo"},
			"/r/demo/foo",
		},
		{
			"user address is kept",
			weburl.GnoURL{Path: "/u/jae"},
			"/u/jae",
		},
		{
			"render path is kept with slashes restored",
			weburl.GnoURL{Path: "/r/demo/boards", Args: "thread/123"},
			"/r/demo/boards:thread/123",
		},
		{
			"source view keeps file name and mode flag",
			weburl.GnoURL{Path: "/r/demo/foo", WebQuery: url.Values{"source": {""}, "file": {"foo.gno"}}},
			"/r/demo/foo$file=foo.gno&source",
		},
		{
			"help view keeps func and flag but masks argument values and send",
			weburl.GnoURL{Path: "/r/demo/foo", WebQuery: url.Values{
				"help":   {""},
				"func":   {"Transfer"},
				"amount": {"100"},
				"dest":   {"g1xyz"},
				".send":  {"5ugnot"},
			}},
			"/r/demo/foo$.send=redacted&amount=redacted&dest=redacted&func=Transfer&help",
		},
		{
			"standard query is dropped",
			weburl.GnoURL{Path: "/r/demo/foo", Query: url.Values{"c": {"d"}}},
			"/r/demo/foo",
		},
		{
			"empty url falls back to root",
			weburl.GnoURL{},
			"/",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := analyticsPath(tc.url); got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}
