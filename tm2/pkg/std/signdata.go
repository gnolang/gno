package std

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/netip"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
)

// SignDataMsgType is the one message of a sign-data document: ADR-036's message
// for off-chain data. A node rebuilds a transaction's sign bytes from the
// decoded transaction, and amino writes every message there as {"@type": ...}:
// a document whose message is {"type", "value"} is no transaction's sign
// document, so a signature over it proves control of an address and can never
// be broadcast. TestSignDataPayloadIsNoTransactionPayload pins this.
const SignDataMsgType = "sign/MsgSignData"

// A sign-data document is a transaction sign document with every field fixed
// but the chain id and the message.
// The chain id stays (ADR-036 wants it empty): a proof made on one network is
// then no proof on another.
type signDataPayload struct {
	ChainID       string        `json:"chain_id"`
	AccountNumber uint64        `json:"account_number"`
	Sequence      uint64        `json:"sequence"`
	Fee           signDocFee    `json:"fee"`
	Msgs          []signDataMsg `json:"msgs"`
	Memo          string        `json:"memo"`
}

// signDataMsg is written in amino's type/value form, which ADR-036 and the
// Ledger Cosmos app use, not as a std.Msg (whose JSON carries "@type").
type signDataMsg struct {
	Type  string        `json:"type"`
	Value signDataValue `json:"value"`
}

// signDataValue is ADR-036's {signer, data}, plus the origin and the purpose
// that bind a proof to one service and one use.
type signDataValue struct {
	Signer  string `json:"signer"`
	Data    string `json:"data"` // base64 of the UTF-8 text
	Origin  string `json:"origin"`
	Purpose string `json:"purpose"`
}

// SignedData is what a sign-data document says: on ChainID, Signer answers the
// service at Origin, for Purpose, with Data.
type SignedData struct {
	ChainID string
	// Origin names the service the proof is for, as scheme://host[:port] (see
	// checkSignDataOrigin). A wallet takes it from the requester, never from the
	// request.
	Origin string
	// Purpose is what the proof is for, as the service names it, such as
	// "username/Reserve" (see checkSignDataPurpose).
	Purpose string
	Signer  crypto.Address
	// Data is a text the user reads before signing, typically the service's
	// challenge.
	Data string
}

// SignDataPayload returns the bytes a key signs to state d: the sign document
// sorted by key, its fee an empty coin list (see signDocFee).
func SignDataPayload(d SignedData) ([]byte, error) {
	switch {
	case d.ChainID == "":
		return nil, errors.New("a sign-data document needs a chain id")
	case d.Signer.IsZero():
		return nil, errors.New("a sign-data document needs a signer")
	case d.Data == "":
		return nil, errors.New("a sign-data document needs data")
	case !utf8.ValidString(d.Data):
		return nil, errors.New("sign-data data must be UTF-8 text")
	}
	if err := checkSignDataOrigin(d.Origin); err != nil {
		return nil, err
	}
	if err := checkSignDataPurpose(d.Purpose); err != nil {
		return nil, err
	}

	return signaturePayload(signDataPayload{
		ChainID: d.ChainID,
		Fee:     signDocFee{Amount: []signDocCoin{}, Gas: "0"},
		Msgs: []signDataMsg{{
			Type: SignDataMsgType,
			Value: signDataValue{
				Signer:  d.Signer.String(),
				Data:    base64.StdEncoding.EncodeToString([]byte(d.Data)),
				Origin:  d.Origin,
				Purpose: d.Purpose,
			},
		}},
	})
}

// maxSignDataOrigin and maxSignDataPurpose bound what a wallet shows the user.
const (
	maxSignDataOrigin  = 256
	maxSignDataPurpose = 64
)

// checkSignDataOrigin checks that origin is written as a browser serializes a
// page's origin: scheme://host[:port], lowercase, the host a DNS name in ASCII
// (an internationalized one in its xn-- form) or an IP address, the port only
// when it is not the scheme's default, and nothing after. One service then has
// one origin, and a verifier compares it as a string.
func checkSignDataOrigin(origin string) error {
	switch {
	case origin == "":
		return errors.New("a sign-data document needs an origin")
	case len(origin) > maxSignDataOrigin:
		return fmt.Errorf("sign-data origin is over %d bytes", maxSignDataOrigin)
	}

	invalid := func(why string) error {
		return fmt.Errorf("sign-data origin %q is not scheme://host[:port]: %s", origin, why)
	}

	scheme, hostport, ok := strings.Cut(origin, "://")
	if !ok || !isOriginScheme(scheme) {
		return invalid("it must start with a lowercase scheme and ://")
	}

	// The port follows the last colon, unless that colon is inside an IPv6 host.
	host, port, hasPort := hostport, "", false
	if i := strings.LastIndexByte(hostport, ':'); i > strings.LastIndexByte(hostport, ']') {
		host, port, hasPort = hostport[:i], hostport[i+1:], true
	}

	if !isOriginHost(host) {
		return invalid("the host must be a lowercase ASCII name or an IP address, with no path, query or fragment after it")
	}

	if hasPort {
		n, err := strconv.ParseUint(port, 10, 16)
		switch {
		case err != nil || port[0] == '0':
			return invalid("the port must be a number from 1 to 65535, without leading zeros")
		case (scheme == "http" && n == 80) || (scheme == "https" && n == 443):
			return invalid("the scheme's default port is left out")
		}
	}

	return nil
}

// isOriginScheme reports whether s is a URI scheme (RFC 3986) in lowercase.
func isOriginScheme(s string) bool {
	if s == "" || s[0] < 'a' || s[0] > 'z' {
		return false
	}
	for i := 1; i < len(s); i++ {
		c := s[i]
		if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '+' || c == '-' || c == '.') {
			return false
		}
	}

	return true
}

// isOriginHost reports whether host is a lowercase DNS name, an IPv4 address,
// or a bracketed IPv6 address, each as a browser writes it.
func isOriginHost(host string) bool {
	if strings.HasPrefix(host, "[") {
		if !strings.HasSuffix(host, "]") {
			return false
		}
		ip, err := netip.ParseAddr(host[1 : len(host)-1])
		return err == nil && ip.Is6() && !ip.Is4In6() && ip.Zone() == "" && ip.String() == host[1:len(host)-1]
	}

	labels := strings.Split(host, ".")
	for _, label := range labels {
		if label == "" || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for i := 0; i < len(label); i++ {
			c := label[i]
			if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '-') {
				return false
			}
		}
	}

	// A browser reads a name whose last label is a number as an IPv4 address,
	// and ParseAddr takes only its dotted-decimal form.
	if strings.Trim(labels[len(labels)-1], "0123456789") == "" {
		_, err := netip.ParseAddr(host)
		return err == nil
	}

	return true
}

// checkSignDataPurpose checks that purpose is a name of ASCII letters, digits
// and . _ / -, at most maxSignDataPurpose bytes. A wallet shows it as it is, so
// it holds no space, line break, control or look-alike character to reshape
// the approval screen with.
func checkSignDataPurpose(purpose string) error {
	switch {
	case purpose == "":
		return errors.New("a sign-data document needs a purpose")
	case len(purpose) > maxSignDataPurpose:
		return fmt.Errorf("sign-data purpose is over %d bytes", maxSignDataPurpose)
	}

	for i := 0; i < len(purpose); i++ {
		c := purpose[i]
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || strings.IndexByte("._/-", c) >= 0) {
			return fmt.Errorf("sign-data purpose %q may only hold ASCII letters, digits and . _ / -", purpose)
		}
	}

	return nil
}

// signDataDoc reads the values of a sign-data document; VerifySignData checks
// the rest of it against the canonical bytes.
type signDataDoc struct {
	ChainID string        `json:"chain_id"`
	Msgs    []signDataMsg `json:"msgs"`
}

// VerifySignData checks that sig, by pubKey, signs the sign-data document doc,
// and that pubKey is the key of the document's signer. Only a secp256k1 key
// signs a proof. It returns what the document says; the caller still checks the
// chain id, origin, purpose and data it expects.
//
// doc may come in any JSON rendering (key order, spacing, string escapes), but
// must otherwise be the canonical document: re-encoded sorted by key and
// compacted, it must equal the bytes rebuilt from its values, and no object may
// repeat a key. Missing or extra fields, numbers for strings, a null for the
// empty fee, non-canonical base64 in the data and a duplicate key are refused,
// so no other document passes for one, whatever parser reads it.
func VerifySignData(doc []byte, pubKey secp256k1.PubKeySecp256k1, sig []byte) (SignedData, error) {
	var d signDataDoc
	if err := unmarshalStrict(doc, &d); err != nil {
		return SignedData{}, fmt.Errorf("unable to read sign-data document, %w", err)
	}
	if len(d.Msgs) != 1 {
		return SignedData{}, errors.New("not a sign-data document")
	}

	signer, err := crypto.AddressFromBech32(d.Msgs[0].Value.Signer)
	if err != nil {
		return SignedData{}, fmt.Errorf("invalid sign-data signer, %w", err)
	}
	data, err := base64.StdEncoding.DecodeString(d.Msgs[0].Value.Data)
	if err != nil {
		return SignedData{}, fmt.Errorf("invalid sign-data data, %w", err)
	}

	signed := SignedData{
		ChainID: d.ChainID,
		Origin:  d.Msgs[0].Value.Origin,
		Purpose: d.Msgs[0].Value.Purpose,
		Signer:  signer,
		Data:    string(data),
	}
	payload, err := SignDataPayload(signed)
	if err != nil {
		return SignedData{}, err
	}

	// doc already parsed, so sortJSON cannot fail.
	if canonical, _ := sortJSON(doc); !bytes.Equal(canonical, payload) {
		return SignedData{}, errors.New("not a sign-data document")
	}

	if pubKey.Address() != signer {
		return SignedData{}, fmt.Errorf("the key belongs to %s, not to the signer %s", pubKey.Address(), signer)
	}
	if !pubKey.VerifyBytes(payload, sig) {
		return SignedData{}, errors.New("invalid sign-data signature")
	}

	return signed, nil
}

// SignDataPubKeyType is the type URL of a sign-data proof's public key.
const SignDataPubKeyType = "/tm.PubKeySecp256k1"

// SignDataProof is a sign-data proof: the signed document, the public key and
// the signature.
type SignDataProof struct {
	Signed    json.RawMessage   `json:"signed"`
	Signature SignDataSignature `json:"signature"`
}

type SignDataSignature struct {
	PubKey    SignDataPubKey `json:"pubKey"`
	Signature string         `json:"signature"` // base64, 64 bytes R||S
}

type SignDataPubKey struct {
	TypeURL string `json:"typeUrl"`
	Value   string `json:"value"` // base64 of the compressed key itself
}

// NewSignDataProof returns the proof that sig, by pubKey, signs payload.
func NewSignDataProof(payload []byte, pubKey secp256k1.PubKeySecp256k1, sig []byte) SignDataProof {
	return SignDataProof{
		Signed: payload,
		Signature: SignDataSignature{
			PubKey: SignDataPubKey{
				TypeURL: SignDataPubKeyType,
				Value:   base64.StdEncoding.EncodeToString(pubKey[:]),
			},
			Signature: base64.StdEncoding.EncodeToString(sig),
		},
	}
}

// VerifySignDataProof checks a proof in SignDataProof's JSON with
// VerifySignData; the caller still checks the chain id, origin, purpose and data
// it expects.
// Its cost grows with the proof's size, so a service bounds what it accepts
// (a proof of a short challenge is under 1 KiB).
//
// A proof stays valid forever, and the same proof can be written in many byte
// forms (the document's rendering, the envelope's base64 and key spelling). A
// service therefore refuses a replay by its challenge, never by the proof's
// bytes: it hands out a fresh, unpredictable challenge naming the session,
// gives it an expiry, and accepts it once. The origin and the purpose keep a proof
// made for one service, or one use, from passing at another.
func VerifySignDataProof(proof []byte) (SignedData, error) {
	var p SignDataProof
	if err := unmarshalStrict(proof, &p); err != nil {
		return SignedData{}, fmt.Errorf("unable to read sign-data proof, %w", err)
	}

	if p.Signature.PubKey.TypeURL != SignDataPubKeyType {
		return SignedData{}, fmt.Errorf("the proof's key is a %q, not a %s", p.Signature.PubKey.TypeURL, SignDataPubKeyType)
	}

	key, err := base64.StdEncoding.DecodeString(p.Signature.PubKey.Value)
	if err != nil || len(key) != secp256k1.PubKeySecp256k1Size {
		return SignedData{}, errors.New("the proof's public key is not a base64 compressed secp256k1 key")
	}

	var pubKey secp256k1.PubKeySecp256k1
	copy(pubKey[:], key)

	sig, err := base64.StdEncoding.DecodeString(p.Signature.Signature)
	if err != nil {
		return SignedData{}, fmt.Errorf("the proof's signature is not base64, %w", err)
	}

	return VerifySignData(p.Signed, pubKey, sig)
}
