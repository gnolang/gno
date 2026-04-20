package components

import "testing"

func TestClassifyView(t *testing.T) {
	cases := []struct {
		name     string
		mode     ViewMode
		view     ViewType
		wantType string
		wantCtx  AnalyticsContext
	}{
		{"source view wins over realm mode", ViewModeRealm, SourceViewType, "source", AnalyticsContextBuilder},
		{"help view wins over realm mode", ViewModeRealm, HelpViewType, "help", AnalyticsContextBuilder},
		{"directory view", ViewModeExplorer, DirectoryViewType, "directory", AnalyticsContextBuilder},
		{"status view", ViewModeExplorer, StatusViewType, "status", AnalyticsContextNeutral},
		{"redirect view", ViewModeExplorer, RedirectViewType, "redirect", AnalyticsContextNeutral},
		{"realm view falls through to mode", ViewModeRealm, RealmViewType, "realm", AnalyticsContextNeutral},
		{"home mode", ViewModeHome, RealmViewType, "home", AnalyticsContextNeutral},
		{"user mode", ViewModeUser, UserViewType, "user", AnalyticsContextBuilder},
		{"package mode", ViewModePackage, RealmViewType, "package", AnalyticsContextBuilder},
		{"explorer mode", ViewModeExplorer, RealmViewType, "explorer", AnalyticsContextBuilder},
		{"unknown view + zero mode", ViewModeExplorer, ViewType("unknown"), "explorer", AnalyticsContextBuilder},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotType, gotCtx := ClassifyView(tc.mode, tc.view)
			if gotType != tc.wantType {
				t.Errorf("pageType: got %q, want %q", gotType, tc.wantType)
			}
			if gotCtx != tc.wantCtx {
				t.Errorf("context: got %q, want %q", gotCtx, tc.wantCtx)
			}
		})
	}
}
