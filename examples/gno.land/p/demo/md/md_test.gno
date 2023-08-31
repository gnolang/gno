package md

import (
	"testing"
)

func TestEscMDURL(t *testing.T) {
	tests := []struct {
		url      string
		expected string
	}{
		{
			url:      "://myurl.com",
			expected: "%3A//myurl.com",
		},
		{
			url:      "://my url.com",
			expected: "%3A//my%20url.com",
		},
		{
			url:      "://my url.com?foo=bar",
			expected: "%3A//my%20url.com%3Ffoo%3Dbar",
		},
	}

	for _, test := range tests {
		actual := EscMDURL(test.url)
		if actual != test.expected {
			t.Errorf("EscMDURL(%q) = %q, expected %q", test.url, actual, test.expected)
		}
	}
}

func TestLink(t *testing.T) {
	tests := []struct {
		text     string
		url      string
		expected string
	}{
		{
			text:     "foo",
			url:      "://myurl.com",
			expected: "[foo](%3A//myurl.com)",
		},
		{
			text:     "bar",
			url:      "://my url.com",
			expected: "[bar](%3A//my%20url.com)",
		},
		{
			text:     "baz",
			url:      "://my url.com?foo=bar",
			expected: "[baz](%3A//my%20url.com%3Ffoo%3Dbar)",
		},
	}

	for _, test := range tests {
		actual := Link(test.text, test.url)
		if actual != test.expected {
			t.Errorf("Link(%q, %q) = %q, expected %q", test.text, test.url, actual, test.expected)
		}
	}
}
