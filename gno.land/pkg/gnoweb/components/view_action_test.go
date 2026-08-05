package components

import (
	"bytes"
	"encoding/base64"
	"html/template"
	"image/png"
	"strings"
	"testing"

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
