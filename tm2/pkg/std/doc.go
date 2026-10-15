package std

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strconv"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/crypto"
)

// SignDoc is the standard object for transactions.
// AccountNumber is a replay-prevention field for the whole account
// (eg. nonce) to prevent the replay of txs after an account has been deleted
// (due to zero balance). Sequence is a replay-prevention field for each transaction
// given a nonce
type SignDoc struct {
	ChainID       string `json:"chain_id" yaml:"chain_id"`
	AccountNumber uint64 `json:"account_number" yaml:"account_number"`
	Sequence      uint64 `json:"sequence" yaml:"sequence"`
	Fee           Fee    `json:"fee" yaml:"fee"`
	Msgs          []Msg  `json:"msgs" yaml:"msgs"`
	Memo          string `json:"memo" yaml:"memo"`
}

// signDocFee is the fee as it appears in the signature payload.
//
// It is deliberately NOT std.Fee. The Ledger Cosmos app -- which gno has no
// app of its own and therefore borrows -- validates the amino sign doc against
// a fixed allowlist of keys, and refuses to sign anything containing a key it
// does not know. Inside "fee" it permits only amount, gas, granter and payer
// (ledger-cosmos, app/src/tx_validate.c).
//
// std.Fee renders as {"gas_wanted":...,"gas_fee":"1000000ugnot"}, so every gno
// transaction -- a send as much as an addpkg -- is refused with "Unexpected
// field" before the device will display anything. Three things differ from what
// the app expects: the key names, the cardinality (one Coin against an array),
// and the numeric type (amino JSON renders these as strings).
type signDocFee struct {
	Amount []signDocCoin `json:"amount"`
	Gas    string        `json:"gas"`
}

// signDocCoin spells out a Coin, because amino renders std.Coin as the single
// string "1000000ugnot" and the allowlist wants an object.
type signDocCoin struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}

// signDocPayload mirrors SignDoc with the fee in the form above. Every other
// field already matches what the app allows, so only the fee is restated.
// TestSignDocPayloadMirrorsSignDoc keeps the two field sets in step.
type signDocPayload struct {
	ChainID       string     `json:"chain_id"`
	AccountNumber uint64     `json:"account_number"`
	Sequence      uint64     `json:"sequence"`
	Fee           signDocFee `json:"fee"`
	Msgs          []Msg      `json:"msgs"`
	Memo          string     `json:"memo"`
}

// GetSignaturePayloadLegacy returns the signature payload with the fee in its
// gas_wanted/gas_fee rendering: amino JSON over SignDoc itself, sorted by key.
//
// Verification accepts this rendering alongside the one GetSignaturePayload
// produces, so that clients building the payload themselves -- wallets, the
// genesis tooling, anything holding a key -- and signatures already written into
// genesis files or chain history keep verifying. See VerifySignaturePayload.
func GetSignaturePayloadLegacy(s SignDoc) ([]byte, error) {
	return signaturePayload(s)
}

// feeAmount renders the fee coin as Cosmos renders a coin list.
//
// A ZERO FEE IS AN EMPTY LIST, not a list holding an empty coin. Cosmos's Coins
// carries no zero entries, and the device DISPLAYS each coin it is given, so
// {"denom":"","amount":"0"} is both wrong and likely to be rejected -- which
// would reintroduce, for zero-fee transactions, exactly the failure this file
// exists to fix. Zero fees are a supported case, not a hypothetical: the ante
// handler branches on GasFee.IsZero(), and every genesis transaction is signed
// with GetSignBytes(chainID, 0, 0).
//
// The slice MUST be non-nil. Amino renders a nil slice as null and an empty one
// as [], and those are different signed bytes; TestSignaturePayloadZeroFeeIsEmptyList
// pins the latter.
func feeAmount(c Coin) []signDocCoin {
	if c.IsZero() {
		return []signDocCoin{}
	}
	return []signDocCoin{{
		Denom:  c.Denom,
		Amount: strconv.FormatInt(c.Amount, 10),
	}}
}

// GetSignaturePayload returns the signature payload for the SignDoc: amino JSON
// of signDocPayload, sorted by key. Every field of s passes through as is except
// the fee, which is restated as the coin list the Ledger Cosmos app parses (see
// signDocFee). The formula for signing is therefore
// sign(sortJSON(aminoJSON(signDocPayload(SignDoc)))).
func GetSignaturePayload(s SignDoc) ([]byte, error) {
	return signaturePayload(signDocPayload{
		ChainID:       s.ChainID,
		AccountNumber: s.AccountNumber,
		Sequence:      s.Sequence,
		Fee: signDocFee{
			Amount: feeAmount(s.Fee.GasFee),
			Gas:    strconv.FormatInt(s.Fee.GasWanted, 10),
		},
		Msgs: s.Msgs,
		Memo: s.Memo,
	})
}

// signaturePayload renders v as amino JSON with its keys sorted, the form every
// signature payload takes: sign(sortJSON(aminoJSON(v))).
func signaturePayload(v any) ([]byte, error) {
	data, err := amino.MarshalJSON(v)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal sign doc, %w", err)
	}

	sortedData, err := sortJSON(data)
	if err != nil {
		return nil, fmt.Errorf("unable to sort payload JSON, %w", err)
	}

	return sortedData, nil
}

// MaxSignDocStringLen is the length, in bytes of the decoded string, above
// which a string value in the signature payload is replaced by its digest in
// the rendering GetSignaturePayloadDigest produces.
//
// CONSENSUS CONSTANT. Moving it changes the signed bytes of every transaction
// holding a string near the boundary, and a node that moved it would reject
// signatures every other node accepts.
//
// 512 is chosen so that no realistic value transfer is ever digested: a bech32
// address is 40 characters and a coin string a handful more, so bank.MsgSend
// renders identically in both renderings and the money path keeps full
// on-device display. That is a statement about realistic transfers rather than
// an invariant of the type -- Coins render as one string, so enough
// denominations in one send will cross any threshold --
// TestDigestNeverAppliesToMsgSend pins where it holds.
const MaxSignDocStringLen = 512

// digestKey is the only key of the object an oversized string becomes.
//
// It begins with '$' DELIBERATELY. Amino JSON object keys come from Go struct
// tags, and no struct tag reachable from a std.Msg starts with '$', so this key
// cannot occur in the rendering of a real message. That is the whole
// disjointness argument between this rendering and the other two; see
// GetSignaturePayloadDigest. TestDigestKeyIsUnreachableFromAmino pins it.
const digestKey = "$sha256"

// Digested names one substitution the digest rendering made. It exists for the
// client, which has to show a signer what the device will not: which field was
// replaced, how long it was, and the digest to compare against the screen.
// Nothing on the consensus path reads it.
type Digested struct {
	// Path locates the field in the payload, e.g.
	// msgs[0].package.files[1].body.
	Path string
	// Length is the byte length of the string that was replaced.
	Length int
	// Sum is the hex digest, exactly as it appears in the payload.
	Sum string
}

// GetSignaturePayloadDigest returns the payload GetSignaturePayload produces
// with every string value longer than MaxSignDocStringLen replaced by the
// object {"$sha256":"<hex>"}. The formula for signing is therefore
// sign(digestPayload(sortJSON(aminoJSON(signDocPayload(SignDoc))))).
//
// WHY THIS EXISTS. The Ledger Cosmos app, which gno borrows for want of an app
// of its own, parses the sign doc on the device and refuses anything over
// FLASH_BUFFER_SIZE, 16384 bytes (ledger-cosmos, app/src/common/tx.c),
// answering APDU 0x6988 before it displays a thing. A realm deployment passes
// that in a single file. The app has no instruction for signing a bare digest
// -- it offers GET_VERSION, GET_ADDR_SECP256K1 and SIGN_SECP256K1 and nothing
// else (app/src/apdu_handler.c) -- so the only bytes a device will ever sign
// are bytes it parsed. Shrinking the rendering is the only lever there is.
//
// WHY STRINGS AND NOT THE WHOLE DOC. What overflows is a handful of long
// values: a file body, a large call argument. The fields that carry the
// authorisation -- creator, package path, function name, file names, fee,
// sequence -- are short and stay readable on the device, so a signer still
// confirms WHERE the code goes and WHO deploys it on trusted hardware, and
// takes only the code text on faith. Hashing the whole document would discard
// all of that to solve the same overflow.
//
// WHY ONLY THE CURRENT FEE SHAPE, with no digest counterpart to the legacy
// rendering: the legacy rendering exists for signatures already made, and this
// one exists for Ledgers, which need the fee shape the app parses. The two
// never meet.
//
// It returns the substitutions it made alongside the payload, because every
// caller that wants one wants the other: the client shows a signer the fields
// their device will not, and asking for them separately would render the
// payload twice.
//
// On a payload with no oversized string this returns THE INPUT ITSELF and no
// substitutions. That is not a coincidence worth testing for, it is how the
// identity is produced, and both callers lean on it: the node to skip a curve
// operation, the client to know the mode changed nothing worth reporting.
func GetSignaturePayloadDigest(s SignDoc) ([]byte, []Digested, error) {
	payload, err := GetSignaturePayload(s)
	if err != nil {
		return nil, nil, err
	}

	var subs []Digested

	digested, _, err := digestPayload(payload, &subs)
	if err != nil {
		return nil, nil, err
	}

	return digested, subs, nil
}

// digestPayload applies the substitution to an already sorted payload, and
// reports whether it replaced anything. When it did not, the payload it returns
// IS payload, so a caller comparing the two renderings needs no byte compare.
//
// subs may be nil, and nil means the caller wants no report. The walk then
// skips the path building and the key ordering, which exist only for that
// report: on the consensus path they were measured at a quarter of the walk's
// time and over half its allocations, all of it discarded.
//
// Nothing is parsed at all unless the payload could hold an oversized string,
// which is what keeps a transaction that cannot have been signed in digest mode
// from paying for the walk. See payloadMayHoldLongString.
func digestPayload(payload []byte, subs *[]Digested) ([]byte, bool, error) {
	if !payloadMayHoldLongString(payload) {
		return payload, false, nil
	}

	return digestWalk(payload, subs)
}

// digestWalk is digestPayload past its gate: it always parses. Separate so that
// the benchmarks can measure the gate against what it saves.
func digestWalk(payload []byte, subs *[]Digested) ([]byte, bool, error) {
	var tree any
	if err := json.Unmarshal(payload, &tree); err != nil {
		return nil, false, fmt.Errorf("unable to parse payload JSON, %w", err)
	}

	substituted := false
	reduced, err := json.Marshal(digestValue(tree, "", subs, &substituted))
	if err != nil {
		return nil, false, fmt.Errorf("unable to marshal digest payload, %w", err)
	}

	// A payload can pass the gate and still hold nothing oversized, since the
	// gate measures escaped lengths. Returning payload keeps the identity exact
	// in that case too.
	if !substituted {
		return payload, false, nil
	}

	return reduced, true, nil
}

// payloadMayHoldLongString reports whether payload holds a JSON string literal
// longer than MaxSignDocStringLen. It is a NECESSARY condition for the digest
// rendering to differ from payload, answered in one pass with no allocation and
// no parsing.
//
// WHY IT IS SOUND IN THE DIRECTION THAT MATTERS. It measures the ESCAPED length
// of each literal, and escaping never shortens a string: a newline occupies two
// bytes as \n, a control character six as \u0000. So a decoded string over the
// threshold always has an escaped form over it too, and a false negative -- the
// one error that would make a node reject a signature its peers accept -- is
// impossible. False positives are possible, cost only the walk that would
// otherwise have run unconditionally, and are caught by digestPayload's
// substituted check. TestPayloadScanHasNoFalseNegatives pins the direction.
//
// WHY IT IS WORTH HAND-WRITING. The alternative is parsing a payload to learn
// that it needed no parsing. Measured on a 1MB transaction made of 7,650 short
// strings -- a shape that can never be signed in digest mode, and the cheapest
// thing an attacker can send a node with a bad signature attached -- the walk
// costs 16ms and 130,000 allocations and this costs 1.6ms and none.
func payloadMayHoldLongString(payload []byte) bool {
	inString, start := false, 0

	for i := 0; i < len(payload); i++ {
		switch payload[i] {
		case '\\':
			if inString {
				i++ // the escaped byte cannot close the string
			}
		case '"':
			if !inString {
				inString, start = true, i+1

				continue
			}
			if i-start > MaxSignDocStringLen {
				return true
			}
			inString = false
		}
	}

	return false
}

// digestValue walks v, replacing oversized strings and reporting through
// substituted whether it replaced any.
//
// When subs is nil nothing is reported, and neither the paths nor the key
// ordering below are computed. Map keys are otherwise visited in sorted order
// so that the reported paths do not depend on map iteration order; the payload
// itself does not depend on the walk at all, since json.Marshal sorts an
// object's keys however it was built.
func digestValue(v any, path string, subs *[]Digested, substituted *bool) any {
	switch t := v.(type) {
	case string:
		if len(t) <= MaxSignDocStringLen {
			return t
		}
		sum := sha256.Sum256([]byte(t))
		hexSum := hex.EncodeToString(sum[:])
		*substituted = true
		if subs != nil {
			*subs = append(*subs, Digested{Path: path, Length: len(t), Sum: hexSum})
		}

		return map[string]any{digestKey: hexSum}
	case map[string]any:
		out := make(map[string]any, len(t))
		if subs == nil {
			for key, value := range t {
				out[key] = digestValue(value, "", nil, substituted)
			}

			return out
		}
		for _, key := range slices.Sorted(maps.Keys(t)) {
			out[key] = digestValue(t[key], joinDigestPath(path, key), subs, substituted)
		}

		return out
	case []any:
		out := make([]any, len(t))
		for i, item := range t {
			itemPath := ""
			if subs != nil {
				itemPath = path + "[" + strconv.Itoa(i) + "]"
			}
			out[i] = digestValue(item, itemPath, subs, substituted)
		}

		return out
	default:
		// Numbers, booleans and null pass through. A number would round-trip
		// through float64 here, which is why nothing in this package renders
		// one: amino writes int64 and uint64 as strings, and sortJSON has
		// always made the same round trip anyway.
		return v
	}
}

func joinDigestPath(path, key string) string {
	if path == "" {
		return key
	}

	return path + "." + key
}

// PayloadRendering names the signature payload rendering a signature was made
// over, as reported by VerifySignaturePayload.
type PayloadRendering uint8

const (
	// PayloadRenderingNone means the signature matched no rendering.
	PayloadRenderingNone PayloadRendering = iota
	// PayloadRenderingCurrent is the amount/gas fee shape the Ledger Cosmos app
	// parses, produced by GetSignaturePayload.
	PayloadRenderingCurrent
	// PayloadRenderingLegacy is the gas_wanted/gas_fee fee shape, produced by
	// GetSignaturePayloadLegacy.
	PayloadRenderingLegacy
	// PayloadRenderingDigest is the current shape with oversized strings
	// replaced by their digests, produced by GetSignaturePayloadDigest.
	PayloadRenderingDigest
)

// VerifySignaturePayload reports which rendering of s, if any, sig is a valid
// signature over by pubKey. An error means no payload could be built to check
// against, which is a malformed sign doc rather than a bad signature.
//
// WHY ALL THREE ARE ACCEPTED. Clients build the signature payload themselves,
// so the rendering cannot change on one side only. A node that took only the
// amount/gas bytes would reject every wallet still producing the other shape,
// and every signature already made over it -- including the ones sitting in
// written genesis files, which cannot be re-signed. A node that took only the
// gas_wanted/gas_fee bytes leaves every Ledger unable to sign at all. A node
// that took neither digest rendering leaves a Ledger unable to sign anything
// over 16KB, which is every realm deployment.
//
// WHY ACCEPTING ALL THREE IS SAFE, and not merely convenient. No two renderings
// can be confused for one another, so each signature still authorises exactly
// one transaction; what widens is the set of acceptable proofs of it, not the
// set of transactions behind a proof.
//
// Current against legacy: one fee object carries gas_wanted and gas_fee, the
// other amount and gas, and those key sets have no member in common. Two JSON
// objects that parse to different key sets are not the same bytes, so no legacy
// rendering of one transaction can equal the current rendering of a DIFFERENT
// one.
//
// Digest against either: the digest rendering differs from the current one only
// where the current one holds a string over MaxSignDocStringLen and the digest
// holds an OBJECT. For the digest rendering of transaction A to equal a full
// rendering of some transaction B, B would need an object where A has a string,
// keyed $sha256 -- a key amino cannot emit, since its keys come from Go struct
// tags. Equality therefore forces A and B to be the same transaction, barring a
// SHA-256 second preimage. Had the substitution produced a STRING, a
// transaction whose field legitimately held that literal would have rendered
// identically to the digest rendering of a different one: one signature, two
// transactions. The object shape is load-bearing.
//
// TestSignaturePayloadEncodingsAreDisjoint and TestDigestRenderingIsDisjoint pin
// this, and it is the property to re-check before any rendering is touched.
//
// Each rendering is computed only when the one before it does not verify, so a
// valid signature over the current rendering pays one encoding and one curve
// operation, as it always did.
//
// BE HONEST ABOUT WHAT A BAD SIGNATURE COSTS, because nothing meters it: the
// cost of rejecting an invalid signature is borne by the node, not the sender.
// A signature matching none of the three pays the legacy encoding, then the
// digest rendering, then up to three curve operations. On a 1MB transaction the
// digest rendering is cheap only because it transforms the payload already in
// hand rather than building one, and because a payload that cannot hold an
// oversized string is never parsed at all -- the two things digestPayload
// exists to do. Without them this function would roughly double what a
// bad-signature flood costs a node; with them the digest arm adds about a
// millisecond and a half to it.
func VerifySignaturePayload(pubKey crypto.PubKey, s SignDoc, sig []byte) (PayloadRendering, error) {
	payload, err := GetSignaturePayload(s)
	if err != nil {
		return PayloadRenderingNone, err
	}

	return VerifySignaturePayloadFrom(pubKey, s, sig, payload)
}

// VerifySignaturePayloadFrom is VerifySignaturePayload for a caller that has
// already built the current rendering of s, and it is where the fallback chain
// actually lives.
//
// IT EXISTS SO THAT THE CHAIN IS WRITTEN ONCE. The ante handler builds the
// payload before it charges gas and outside its simulate gate, so it cannot
// call a function that builds its own; it used to spell the whole chain out
// again instead, and the copy drifted -- its digest arm rebuilt the payload
// from the transaction while this one transformed the payload in hand, which on
// a 1MB transaction cost the consensus path an extra encoding worth several
// curve operations per rejected signature. Taking the payload as an argument
// removes the reason for the copy without moving the gas charge.
func VerifySignaturePayloadFrom(pubKey crypto.PubKey, s SignDoc, sig, payload []byte) (PayloadRendering, error) {
	if pubKey.VerifyBytes(payload, sig) {
		return PayloadRenderingCurrent, nil
	}

	legacy, err := GetSignaturePayloadLegacy(s)
	if err != nil {
		return PayloadRenderingNone, err
	}
	if pubKey.VerifyBytes(legacy, sig) {
		return PayloadRenderingLegacy, nil
	}

	// substituted is false exactly when the digest rendering is the payload
	// above, which is every transaction that cannot have been signed in digest
	// mode. Verifying it a second time would prove nothing.
	digested, substituted, err := digestPayload(payload, nil)
	if err != nil {
		return PayloadRenderingNone, err
	}
	if substituted && pubKey.VerifyBytes(digested, sig) {
		return PayloadRenderingDigest, nil
	}

	return PayloadRenderingNone, nil
}

// Signature represents a wrapped signature of a transaction
type Signature struct {
	PubKey    crypto.PubKey `json:"pub_key" yaml:"pub_key"` // optional
	Signature []byte        `json:"signature" yaml:"signature"`
	// SessionAddr identifies a session account for delegated signing.
	// Zero-value means a master-key signature. When non-zero, the AnteHandler
	// loads the session account at /a/<signer>/s/<SessionAddr> for verification.
	// Amino binary and JSON both skip the field when the address is zero
	// ([20]byte{}), so master-signed txs pay no wire-size overhead.
	SessionAddr crypto.Address `json:"session_addr,omitempty" yaml:"session_addr,omitempty"`
}
