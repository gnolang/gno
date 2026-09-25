package components

import (
	"bytes"
	"encoding/base64"
	"html/template"
	"image/png"
	"strings"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
	"github.com/gnolang/gno/gnovm/pkg/doc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func helpFixture() (HelpData, HelpFunction) {
	fn := HelpFunction{JSONFunc: &doc.JSONFunc{
		Name:   "Transfer",
		Params: []*doc.JSONField{{Name: "to"}, {Name: "amount"}},
	}}
	data := HelpData{
		SelectedFunc: "Transfer",
		SelectedArgs: map[string]string{"to": "g1abc", "amount": "42"},
		RealmName:    "bank",
		PkgPath:      "gno.land/r/demo/bank",
		Domain:       "gno.land",
		Origin:       "https://gno.land",
		Functions:    []HelpFunction{fn},
	}
	return data, fn
}

func TestBuildHelpQR_EncodesTheHelpURL(t *testing.T) {
	t.Parallel()

	data, fn := helpFixture()
	qrFunc, ok := funcMap["buildHelpQR"].(func(HelpData, HelpFunction) (template.URL, error))
	require.True(t, ok, "buildHelpQR must be registered with the expected signature")

	uri, err := qrFunc(data, fn)
	require.NoError(t, err)

	raw := string(uri)
	require.True(t, strings.HasPrefix(raw, "data:image/png;base64,"), "got %.40q", raw)

	payload, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(raw, "data:image/png;base64,"))
	require.NoError(t, err)

	img, err := png.Decode(bytes.NewReader(payload))
	require.NoError(t, err, "the payload must be a decodable PNG")
	assert.GreaterOrEqual(t, img.Bounds().Dx(), 128, "QR must be large enough to scan")
	assert.Equal(t, img.Bounds().Dx(), img.Bounds().Dy(), "QR must be square")
}

func TestBuildHelpQR_IsDeterministic(t *testing.T) {
	t.Parallel()

	data, fn := helpFixture()
	qrFunc := funcMap["buildHelpQR"].(func(HelpData, HelpFunction) (template.URL, error))

	first, err := qrFunc(data, fn)
	require.NoError(t, err)
	second, err := qrFunc(data, fn)
	require.NoError(t, err)
	assert.Equal(t, first, second)

	// A different argument must produce a different code.
	other := data
	other.SelectedArgs = map[string]string{"to": "g1xyz", "amount": "42"}
	third, err := qrFunc(other, fn)
	require.NoError(t, err)
	assert.NotEqual(t, first, third)
}

func TestHelpView_RendersQRPanel(t *testing.T) {
	t.Parallel()

	data, _ := helpFixture()
	var buf bytes.Buffer
	require.NoError(t, HelpView(data).Render(&buf))

	out := buf.String()
	assert.Contains(t, out, `id="qr-Transfer"`)
	assert.Contains(t, out, "data:image/png;base64,")
	assert.Contains(t, out, `data-action-function-target="qr-anchor"`)
}

// postFixture is helpFixture with a two-argument function, so the URL tests
// vary only the args.
func postFixture(args map[string]string) (HelpData, HelpFunction) {
	fn := HelpFunction{JSONFunc: &doc.JSONFunc{
		Name:   "Post",
		Params: []*doc.JSONField{{Name: "author"}, {Name: "body"}},
	}}
	return HelpData{
		SelectedFunc: "Post",
		SelectedArgs: args,
		PkgPath:      "gno.land/r/demo/board",
		Domain:       "gno.land",
		Origin:       "https://gno.land",
		Functions:    []HelpFunction{fn},
	}, fn
}

// Args reach buildHelpURL already decoded, so the URL it builds must survive
// being parsed again — otherwise the form action, the copied link and the QR
// describe a different call than the page shows.
func TestBuildHelpURL_ArgsSurviveAReparse(t *testing.T) {
	t.Parallel()

	for name, author := range map[string]string{
		"separator":   "a&b=evil", // injected a parameter
		"fragment":    "tag#frag", // truncated the rest of the URL
		"percent":     "100%",     // made the URL unparseable
		"space":       "a b",      // raw space in a URL
		"plus":        "a+b",      // must not come back as a space
		"unicode":     "héllo",
		"empty":       "",
		"already_enc": "%26",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			data, fn := postFixture(map[string]string{"author": author, "body": "hello"})
			parsed, err := weburl.Parse(buildHelpURL(data, fn))
			require.NoError(t, err)
			assert.Equal(t, "Post", parsed.WebQuery.Get("func"))
			assert.Equal(t, author, parsed.WebQuery.Get("author"))
			assert.Equal(t, "hello", parsed.WebQuery.Get("body"))
		})
	}
}

// A reparse alone cannot pin the spelling: url.ParseQuery decodes `+` and
// `%20` alike. The frontend writes `%20`, so assert on the bytes.
func TestBuildHelpURL_SpellsSpaceLikeTheFrontend(t *testing.T) {
	t.Parallel()

	data, fn := postFixture(map[string]string{"author": "a b"})
	assert.Contains(t, buildHelpURL(data, fn), "author=a%20b")
}

// Pins HelpView's own contract, not the handler's output: GetHelpView narrows to
// the selected function, so it never renders this two-function shape. What bites
// in production is the `$help` index, where nothing is selected and every
// function would otherwise encode a ~3ms PNG.
func TestHelpView_RendersQROnlyForTheSelectedFunction(t *testing.T) {
	t.Parallel()

	other := HelpFunction{JSONFunc: &doc.JSONFunc{
		Name:   "Render",
		Params: []*doc.JSONField{{Name: "path"}},
	}}
	data, fn := helpFixture()
	data.Functions = []HelpFunction{fn, other}

	var buf bytes.Buffer
	require.NoError(t, HelpView(data).Render(&buf))

	out := buf.String()
	assert.Contains(t, out, `id="qr-Transfer"`, "the selected function keeps its QR")
	assert.NotContains(t, out, `id="qr-Render"`, "an unselected function must not encode one")
	assert.Equal(t, 1, strings.Count(out, "data:image/png;base64,"))
	// The button still points at the other function, which selects it on arrival.
	assert.Contains(t, out, "func=Render")
}
