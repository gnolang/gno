package vm

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/ledger"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The transaction the digest rendering exists for. These pins live here rather
// than beside the rendering because tm2 cannot import gno.land: std sees no
// MsgAddPackage, so nothing there can show what the mode does to the message
// that motivated it.

// fileBody is a file body of a realistic size for a realm, well past what the
// device will parse.
func fileBody() string {
	return "package foo\n" + strings.Repeat("// a line of source code\n", 800)
}

func addPackageDoc(t *testing.T, body string) std.SignDoc {
	t.Helper()

	return std.SignDoc{
		ChainID:       "dev",
		AccountNumber: 42,
		Sequence:      7,
		Fee:           std.NewFee(200000, std.Coin{Denom: "ugnot", Amount: 1000000}),
		Msgs: []std.Msg{MsgAddPackage{
			Creator: crypto.AddressFromPreimage([]byte("creator")),
			Package: &std.MemPackage{
				Name: "foo",
				Path: "gno.land/r/albttx/foo",
				Files: []*std.MemFile{
					{Name: "gnomod.toml", Body: "module = \"gno.land/r/albttx/foo\"\ngno = \"0.9\"\n"},
					{Name: "foo.gno", Body: body},
				},
			},
		}},
		Memo: "deploy foo",
	}
}

// WHAT THE DEVICE ENDS UP SEEING, pinned byte for byte because it is consensus
// data on one side and the signer's only evidence on the other.
//
// Read what survives: the chain, the account, the sequence, the fee, the
// creator, the package path, the name of every file and the whole of
// gnomod.toml, which is short enough to pass through. One value is opaque. That
// is the difference between this and hashing the document.
func TestAddPackageDigestRenderingPinned(t *testing.T) {
	t.Parallel()

	body := fileBody()
	sum := sha256.Sum256([]byte(body))
	want := `{"account_number":"42","chain_id":"dev","fee":{"amount":[{"amount":"1000000",` +
		`"denom":"ugnot"}],"gas":"200000"},"memo":"deploy foo","msgs":[{"@type":"/vm.m_addpkg",` +
		`"creator":"g1h34lmpywh4upnjdg90cjf4j70aee6z8qpqxv3c","max_deposit":"","package":` +
		`{"files":[{"body":"module = \"gno.land/r/albttx/foo\"\ngno = \"0.9\"\n",` +
		`"name":"gnomod.toml"},{"body":{"$sha256":"` + hex.EncodeToString(sum[:]) + `"},` +
		`"name":"foo.gno"}],"name":"foo","path":"gno.land/r/albttx/foo"},"send":""}],"sequence":"7"}`

	payload, _, err := std.GetSignaturePayloadDigest(addPackageDoc(t, body))
	require.NoError(t, err)
	assert.Equal(t, want, string(payload), "the digest rendering changed -- this is a consensus change")
}

// The whole point: the full rendering is past what the device will parse, and
// the digest rendering is not. A regression here is the mode silently ceasing
// to solve the problem it was added for.
func TestAddPackageDigestFitsTheDevice(t *testing.T) {
	t.Parallel()

	doc := addPackageDoc(t, fileBody())

	full, err := std.GetSignaturePayload(doc)
	require.NoError(t, err)
	require.Greater(t, len(full), ledger.MaxPayloadSize,
		"this fixture no longer overflows the device, so it proves nothing")

	digested, _, err := std.GetSignaturePayloadDigest(doc)
	require.NoError(t, err)
	assert.LessOrEqual(t, len(digested), ledger.MaxPayloadSize,
		"the digest rendering of a realm deployment still does not fit the device")
}

// Only the file bodies are given up. A signer confirms on trusted hardware
// WHERE the code goes and WHO deploys it, and takes only the code text on
// faith; if a path or a creator ever starts being digested, the mode stops
// being worth its cost.
func TestAddPackageDigestsOnlyFileBodies(t *testing.T) {
	t.Parallel()

	_, fields, err := std.GetSignaturePayloadDigest(addPackageDoc(t, fileBody()))
	require.NoError(t, err)
	require.Len(t, fields, 1)
	assert.Equal(t, "msgs[0].package.files[1].body", fields[0].Path)
	assert.Equal(t, len(fileBody()), fields[0].Length)
}

// A deployment small enough for the device renders identically in both modes,
// so signing a small package in digest mode costs the signer no display and
// the node no extra verification.
func TestSmallAddPackageIsUntouched(t *testing.T) {
	t.Parallel()

	doc := addPackageDoc(t, "package foo\n\nfunc Hello() string { return \"hello\" }\n")

	full, err := std.GetSignaturePayload(doc)
	require.NoError(t, err)
	digested, _, err := std.GetSignaturePayloadDigest(doc)
	require.NoError(t, err)

	assert.Equal(t, string(full), string(digested))
}
