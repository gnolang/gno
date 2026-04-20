package components

import "testing"

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
		{"package mode", ViewModePackage, RealmViewType, "package"},
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
