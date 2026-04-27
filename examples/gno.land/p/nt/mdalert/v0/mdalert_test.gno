package mdalert_test

import (
	"testing"

	"gno.land/p/nt/mdalert/v0"
)

func TestAlertString(t *testing.T) {
	tests := []struct {
		name     string
		expected string
		alert    mdalert.Alert
	}{
		{
			name:     "alert",
			expected: "> [!INFO] Title\n> Message\n",
			alert:    mdalert.New(mdalert.TypeInfo, "Title", "Message", false),
		},
		{
			name:     "alert with empty title",
			expected: "> [!INFO]\n> Message\n",
			alert:    mdalert.New(mdalert.TypeInfo, "", "Message", false),
		},
		{
			name:     "alert multiline",
			expected: "> [!INFO]\n> Line1\n> Line2\n",
			alert:    mdalert.New(mdalert.TypeInfo, "", "Line1\nLine2", false),
		},
		{
			name:     "folded alert",
			expected: "> [!INFO]- Title\n> Message\n",
			alert:    mdalert.New(mdalert.TypeInfo, "Title", "Message", true),
		},
		{
			name: "empty alert",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.alert.String()
			if got != tt.expected {
				t.Errorf("Got:  %q\nWant: %q", got, tt.expected)
			}
		})
	}
}

func TestHelpers(t *testing.T) {
	tests := []struct {
		name     string
		expected string
		fn       func() string
	}{
		// CAUTION
		{"caution", "> [!CAUTION] Title\n> Message\n", func() string {
			return mdalert.Caution("Title", "Message")
		}},
		{"caution with empty title", "> [!CAUTION]\n> Message\n", func() string {
			return mdalert.Caution("", "Message")
		}},
		{"caution multiline", "> [!CAUTION] Title\n> Line1\n> Line2\n", func() string {
			return mdalert.Caution("Title", "Line1\nLine2")
		}},
		{"caution formatted", "> [!CAUTION] Title\n> Message\n", func() string {
			return mdalert.Cautionf("Title", "%s", "Message")
		}},
		{"caution formatted with empty title", "> [!CAUTION]\n> Message\n", func() string {
			return mdalert.Cautionf("", "%s", "Message")
		}},
		{"caution formatted multiline", "> [!CAUTION] Title\n> Line1\n> Line2\n", func() string {
			return mdalert.Cautionf("Title", "%s\n%s", "Line1", "Line2")
		}},

		// INFO
		{"info", "> [!INFO] Title\n> Message\n", func() string {
			return mdalert.Info("Title", "Message")
		}},
		{"info with empty title", "> [!INFO]\n> Message\n", func() string {
			return mdalert.Info("", "Message")
		}},
		{"info multiline", "> [!INFO] Title\n> Line1\n> Line2\n", func() string {
			return mdalert.Info("Title", "Line1\nLine2")
		}},
		{"info formatted", "> [!INFO] Title\n> Message\n", func() string {
			return mdalert.Infof("Title", "%s", "Message")
		}},
		{"info formatted with empty title", "> [!INFO]\n> Message\n", func() string {
			return mdalert.Infof("", "%s", "Message")
		}},
		{"info formatted multiline", "> [!INFO] Title\n> Line1\n> Line2\n", func() string {
			return mdalert.Infof("Title", "%s\n%s", "Line1", "Line2")
		}},

		// NOTE
		{"note", "> [!NOTE] Title\n> Message\n", func() string {
			return mdalert.Note("Title", "Message")
		}},
		{"note with empty title", "> [!NOTE]\n> Message\n", func() string {
			return mdalert.Note("", "Message")
		}},
		{"note multiline", "> [!NOTE] Title\n> Line1\n> Line2\n", func() string {
			return mdalert.Note("Title", "Line1\nLine2")
		}},
		{"note formatted", "> [!NOTE] Title\n> Message\n", func() string {
			return mdalert.Notef("Title", "%s", "Message")
		}},
		{"note formatted with empty title", "> [!NOTE]\n> Message\n", func() string {
			return mdalert.Notef("", "%s", "Message")
		}},
		{"note formatted multiline", "> [!NOTE] Title\n> Line1\n> Line2\n", func() string {
			return mdalert.Notef("Title", "%s\n%s", "Line1", "Line2")
		}},

		// SUCCESS
		{"success", "> [!SUCCESS] Title\n> Message\n", func() string {
			return mdalert.Success("Title", "Message")
		}},
		{"success with empty title", "> [!SUCCESS]\n> Message\n", func() string {
			return mdalert.Success("", "Message")
		}},
		{"success multiline", "> [!SUCCESS] Title\n> Line1\n> Line2\n", func() string {
			return mdalert.Success("Title", "Line1\nLine2")
		}},
		{"success formatted", "> [!SUCCESS] Title\n> Message\n", func() string {
			return mdalert.Successf("Title", "%s", "Message")
		}},
		{"success formatted with empty title", "> [!SUCCESS]\n> Message\n", func() string {
			return mdalert.Successf("", "%s", "Message")
		}},
		{"success formatted multiline", "> [!SUCCESS] Title\n> Line1\n> Line2\n", func() string {
			return mdalert.Successf("Title", "%s\n%s", "Line1", "Line2")
		}},

		// TIP
		{"tip", "> [!TIP] Title\n> Message\n", func() string {
			return mdalert.Tip("Title", "Message")
		}},
		{"tip with empty title", "> [!TIP]\n> Message\n", func() string {
			return mdalert.Tip("", "Message")
		}},
		{"tip multiline", "> [!TIP] Title\n> Line1\n> Line2\n", func() string {
			return mdalert.Tip("Title", "Line1\nLine2")
		}},
		{"tip formatted", "> [!TIP] Title\n> Message\n", func() string {
			return mdalert.Tipf("Title", "%s", "Message")
		}},
		{"tip formatted with empty title", "> [!TIP]\n> Message\n", func() string {
			return mdalert.Tipf("", "%s", "Message")
		}},
		{"tip formatted multiline", "> [!TIP] Title\n> Line1\n> Line2\n", func() string {
			return mdalert.Tipf("Title", "%s\n%s", "Line1", "Line2")
		}},

		// WARNING
		{"warning", "> [!WARNING] Title\n> Message\n", func() string {
			return mdalert.Warning("Title", "Message")
		}},
		{"warning with empty title", "> [!WARNING]\n> Message\n", func() string {
			return mdalert.Warning("", "Message")
		}},
		{"warning multiline", "> [!WARNING] Title\n> Line1\n> Line2\n", func() string {
			return mdalert.Warning("Title", "Line1\nLine2")
		}},
		{"warning formatted", "> [!WARNING] Title\n> Message\n", func() string {
			return mdalert.Warningf("Title", "%s", "Message")
		}},
		{"warning formatted with empty title", "> [!WARNING]\n> Message\n", func() string {
			return mdalert.Warningf("", "%s", "Message")
		}},
		{"warning formatted multiline", "> [!WARNING] Title\n> Line1\n> Line2\n", func() string {
			return mdalert.Warningf("Title", "%s\n%s", "Line1", "Line2")
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.fn()
			if got != tt.expected {
				t.Errorf("Got:  %q\nWant: %q", got, tt.expected)
			}
		})
	}
}
