package std

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"unicode/utf8"

	"github.com/gnolang/gno/tm2/pkg/crypto"
)

// SignDataMsgType is the one message of a sign-data document: ADR-036's message
// for off-chain data. Nothing registers it with amino, so no node can decode a
// transaction carrying it: a signature over the document proves control of an
// address and can never be broadcast as a transaction.
const SignDataMsgType = "sign/MsgSignData"

// A sign-data document is a transaction sign document with every field fixed
// but the chain id, the signer and the data, as Adena's SignArbitrary builds it.
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

type signDataValue struct {
	Signer string `json:"signer"`
	Data   string `json:"data"` // base64 of the UTF-8 text
}

// SignedData is what a verified sign-data document says.
type SignedData struct {
	ChainID string
	Signer  crypto.Address
	Data    string
}

// SignDataPayload returns the bytes a key signs to prove, on chainID, that it
// controls signer, with data (a text the user reads before signing, typically a
// verifier's challenge). The bytes are those of Adena's SignArbitrary: the sign
// document sorted by key, its fee an empty coin list (see signDocFee).
func SignDataPayload(chainID string, signer crypto.Address, data string) ([]byte, error) {
	if chainID == "" {
		return nil, errors.New("a sign-data document needs a chain id")
	}
	if signer.IsZero() {
		return nil, errors.New("a sign-data document needs a signer")
	}
	if data == "" {
		return nil, errors.New("a sign-data document needs data")
	}
	if !utf8.ValidString(data) {
		return nil, errors.New("sign-data data must be UTF-8 text")
	}

	return signaturePayload(signDataPayload{
		ChainID: chainID,
		Fee:     signDocFee{Amount: []signDocCoin{}, Gas: "0"},
		Msgs: []signDataMsg{{
			Type: SignDataMsgType,
			Value: signDataValue{
				Signer: signer.String(),
				Data:   base64.StdEncoding.EncodeToString([]byte(data)),
			},
		}},
	})
}

// signDataDoc reads a sign-data document as a verifier receives it, in any key
// order or spacing. Numbers stay strings, as amino JSON writes them.
type signDataDoc struct {
	ChainID       string `json:"chain_id"`
	AccountNumber string `json:"account_number"`
	Sequence      string `json:"sequence"`
	Fee           struct {
		Amount []json.RawMessage `json:"amount"`
		Gas    string            `json:"gas"`
	} `json:"fee"`
	Msgs []signDataMsg `json:"msgs"`
	Memo string        `json:"memo"`
}

// VerifySignData checks that sig, by pubKey, signs the sign-data document doc,
// and that pubKey is the key of the document's signer. It returns what the
// document says; the caller still checks the chain id and the data it expects.
//
// The signature is checked over the document's canonical bytes, rebuilt from
// its values, so doc may come in any key order or spacing. Its shape is fixed:
// anything but zeros, an empty fee, no memo and one sign/MsgSignData message is
// refused, so no other document passes for one.
func VerifySignData(doc []byte, pubKey crypto.PubKey, sig []byte) (SignedData, error) {
	var d signDataDoc
	dec := json.NewDecoder(bytes.NewReader(doc))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&d); err != nil {
		return SignedData{}, fmt.Errorf("unable to read sign-data document, %w", err)
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return SignedData{}, errors.New("unexpected data after the sign-data document")
	}

	if d.AccountNumber != "0" || d.Sequence != "0" || d.Fee.Gas != "0" ||
		len(d.Fee.Amount) != 0 || d.Memo != "" ||
		len(d.Msgs) != 1 || d.Msgs[0].Type != SignDataMsgType {
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

	payload, err := SignDataPayload(d.ChainID, signer, string(data))
	if err != nil {
		return SignedData{}, err
	}

	if pubKey == nil {
		return SignedData{}, errors.New("no public key to verify the sign-data document with")
	}
	if pubKey.Address() != signer {
		return SignedData{}, fmt.Errorf("the key belongs to %s, not to the signer %s", pubKey.Address(), signer)
	}
	if !pubKey.VerifyBytes(payload, sig) {
		return SignedData{}, errors.New("invalid sign-data signature")
	}

	return SignedData{ChainID: d.ChainID, Signer: signer, Data: string(data)}, nil
}
