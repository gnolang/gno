package components

import (
	"html/template"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBannerVariantClass(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		variant BannerVariant
		want    string
	}{
		{"explicit warning", BannerWarning, "b-banner--warning"},
		{"explicit caution", BannerCaution, "b-banner--caution"},
		{"empty falls back to brand", "", "b-banner--brand"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			banner, err := NewBannerData("hello", BannerOptions{Variant: tc.variant})
			require.NoError(t, err)
			assert.Equal(t, tc.want, banner.VariantClass())
		})
	}
}

// TestBannerVariantClassesAreLiteral guards the purgecss contract: every
// modifier class must appear verbatim in Go source, because purgecss extracts
// whole tokens and drops any class it cannot find. A refactor that builds the
// name as "b-banner--"+variant would still pass VariantClass tests and then
// ship an unstyled banner in production.
func TestBannerVariantClassesAreLiteral(t *testing.T) {
	t.Parallel()

	for variant, class := range bannerVariantClass {
		assert.Equal(t, "b-banner--"+string(variant), class,
			"variant %q must map to its own literal modifier class", variant)
		assert.NotContains(t, class, " ", "class names carry no whitespace")
	}
}

func TestNewBannerDataRejectsBadStyle(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		opts BannerOptions
	}{
		{"unknown variant", BannerOptions{Variant: "chartreuse"}},
		{"color with a quote", BannerOptions{Color: `red" onload="alert(1)`}},
		{"color with a semicolon", BannerOptions{Color: "red;position:fixed"}},
		{"color with whitespace", BannerOptions{Color: "rgb(1 2 3)"}},
		{"hex of the wrong length", BannerOptions{Color: "#12345"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := NewBannerData("hello", tc.opts)
			assert.Error(t, err)
		})
	}
}

func TestBannerColor(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		color string
		want  string
	}{
		{"unset", "", ""},
		{"short hex", "#f80", "#f80"},
		{"long hex", "#ff8800", "#ff8800"},
		{"hex with alpha", "#ff8800cc", "#ff8800cc"},
		{"keyword", "rebeccapurple", "rebeccapurple"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			banner, err := NewBannerData("hello", BannerOptions{Color: tc.color})
			require.NoError(t, err)
			assert.Equal(t, tc.want, banner.Color())
		})
	}
}

// TestBannerColorSurvivesCSSEscaping pins the assumption the template relies
// on: html/template escapes the color in CSS context, and every color that
// passes validation comes out the other side unchanged. If the stdlib ever
// starts neutering one of these, the banner would silently lose its color, so
// fail here instead.
func TestBannerColorSurvivesCSSEscaping(t *testing.T) {
	t.Parallel()

	tmpl := template.Must(template.New("t").Parse(`<div style="--banner-bg:{{ . }}">`))

	for _, color := range []string{"#f80", "#ff8800", "#ff8800cc", "rebeccapurple"} {
		require.True(t, ValidBannerColor(color))

		var buf strings.Builder
		require.NoError(t, tmpl.Execute(&buf, color))
		assert.Equal(t, `<div style="--banner-bg:`+color+`">`, buf.String())
	}
}

// TestBannerColorEscapesHostileInput is belt and braces: these values never
// reach a banner because NewBannerData rejects them, but if validation were
// ever loosened, html/template must still keep them inside the attribute.
func TestBannerColorEscapesHostileInput(t *testing.T) {
	t.Parallel()

	tmpl := template.Must(template.New("t").Parse(`<div style="--banner-bg:{{ . }}">`))

	for _, hostile := range []string{`red" onload="alert(1)`, "red;position:fixed", `red;}</style><script>`} {
		require.False(t, ValidBannerColor(hostile), "%q must be rejected up front", hostile)

		var buf strings.Builder
		require.NoError(t, tmpl.Execute(&buf, hostile))
		out := buf.String()
		assert.NotContains(t, out, "<script")
		assert.NotContains(t, out, "onload")
		assert.Equal(t, 2, strings.Count(out, `"`), "output must carry exactly the two attribute quotes: %s", out)
	}
}
