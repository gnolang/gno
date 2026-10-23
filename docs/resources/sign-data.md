# Sign-data: proving control of an address off-chain

A service that needs to know an address belongs to its visitor, to reserve a
name, link an account or open a session, hands out a text, its challenge, and
asks for a signature over it. A sign-data proof is that signature, over a
document that binds it to one chain, one service and one use, and that can never
be broadcast as a transaction.

This page specifies the document, the proof that carries it, and what a wallet
and a verifier each check. `std.SignDataPayload` builds the document's bytes,
`std.VerifySignDataProof` checks a proof, and `gnokey sign-data` and
`gnokey verify-data` do both from the command line (see
[Interacting with gnokey](../users/interact-with-gnokey.md#proving-you-control-an-address)).

## The document

The signer signs a transaction sign document with every field fixed but the
chain id and its single message, `sign/MsgSignData`:

```json
{"account_number":"0","chain_id":"<chain id>","fee":{"amount":[],"gas":"0"},"memo":"",
 "msgs":[{"type":"sign/MsgSignData","value":{"data":"<base64 of the text>",
  "origin":"<origin>","purpose":"<purpose>","signer":"<address>"}}],"sequence":"0"}
```

The bytes signed are that JSON with its keys sorted, no spaces, and `<`, `>`
and `&` escaped as `<`, `>` and `&`. No other field may appear,
and none may be left out.

| Field | Holds |
|-------|-------|
| `chain_id` | the chain the proof is for: a proof made on one network is none on another |
| `signer` | the address, in bech32 |
| `data` | the text, standard base64 with padding: non-empty UTF-8 |
| `origin` | the service the proof is for (see below) |
| `purpose` | what the proof is for, as the service names it (see below) |

The message is written as `{"type", "value"}`. A node rebuilds a transaction's
sign bytes from the decoded transaction, and writes every message there as
`{"@type": ...}`, so no transaction's sign bytes are a sign-data document's. The
root keys are a transaction sign document's on purpose: a hardware wallet that
signs transactions in this JSON signs a proof the same way, and `origin` and
`purpose` sit inside the message rather than at the root so that such a device
does not refuse the document.

### `origin`

The origin is written as a browser serializes a web page's origin
(`location.origin`), so that one service has one origin and a verifier compares
it as a string:

- `scheme://host[:port]`, and nothing after: no path, not even `/`, no query,
  no fragment, no user info;
- the scheme in lowercase;
- the host a lowercase DNS name in ASCII, an internationalized one in its
  `xn--` form, an IPv4 address in dotted decimal, or an IPv6 address in
  brackets, in its shortest form;
- the port only when it is not the scheme's default (443 for `https`, 80 for
  `http`), without leading zeros;
- at most 256 bytes.

`https://gno.land`, `http://localhost:8888` and `http://[::1]:26657` are
origins; `https://gno.land/`, `https://Gno.land` and `https://gno.land:443` are
not, and are refused rather than rewritten.

**A wallet takes the origin from the requester, never from the request.** For a
web page that is the page's origin as the browser reports it. A requester that
is not a web page needs an equivalent the wallet establishes itself;
[GnoConnect](gnoconnect.md) defines it for its launch links. A proof whose
origin the requester chose binds nothing.

### `purpose`

The purpose names the use, such as `auth/SignIn` or `username/Reserve`. It is
1 to 64 ASCII letters, digits, and `.`, `_`, `/` or `-`. A wallet shows it as
it is, so it holds no space, line break, control or look-alike character with
which a requester could reshape the approval screen. A service gives each use a
purpose of its own: a proof made to sign in is then no proof for a transfer of
a name.

## The proof

A proof is the document, as a JSON object, with the signer's public key and the
signature:

```json
{
  "signed": { ...the document... },
  "signature": {
    "pubKey": {"typeUrl": "/tm.PubKeySecp256k1", "value": "<base64 of the compressed key>"},
    "signature": "<base64 of the 64 bytes R||S>"
  }
}
```

Only a secp256k1 key signs a proof.

## Signing

Before signing, a wallet shows the user the origin, the purpose, the chain, the
signing address, and the whole text decoded as UTF-8 (not its base64), and
says that nothing goes on chain. The text is the requester's, so the wallet
shows its control and bidirectional-override characters visibly rather than
letting them act. It signs only a document built from those
values, with the origin it established itself.

## Verifying

A verifier rebuilds the canonical bytes from the values of `signed` and checks
the signature over them, so that `signed` may reach it in any key order,
spacing or string escaping. It then checks that:

- `signed`, sorted and compacted, equals the rebuilt bytes, and no object in
  the proof repeats a key: nothing but the canonical document passes for one;
- the key is `signer`'s;
- `chain_id`, `origin` and `purpose` are the ones it expects;
- `data` is a challenge it handed out, has not expired and has not been used.

`std.VerifySignDataProof` makes the first two checks and returns the values for
the service to compare. Checking the signature alone is not enough: any
signature the key ever made would then pass for a proof.

A proof stays valid forever, and one proof can be written in many byte forms.
A service therefore refuses a replay by its challenge, never by the proof's
bytes: it hands out a fresh, unpredictable challenge, gives it an expiry, and
accepts it once.

## The earlier format

An earlier version of this format signed ADR-036's `{signer, data}` with the
chain id, and nothing else, in the message. Its proofs are refused by
`std.VerifySignDataProof`, since a proof without `origin` and `purpose` binds
neither the service nor the use.

The two can be told apart: only the current document's message carries
`origin`. While wallets still produce the earlier one, a service that has to
accept it does so knowingly, writing its own origin and the use into the
challenge text and checking them there, and stops once its users' wallets have
moved on.

## Test vectors

`tm2/pkg/std/signdata_test.go` pins the bytes and the signatures of a
`gnoland-1` proof, for the origin `https://gno.land` and the purpose
`auth/SignIn`, from the key of gnoland's default test mnemonic. The signatures
come from another implementation, `@cosmjs/amino`, so that an implementation
checked against them agrees with more than itself.
