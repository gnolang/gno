# Support the Gno.land Ledger app in gnokey

## Context

`gnokey` signs with a Ledger device through the Cosmos app (`CLA 0x55`, via
`cosmos/ledger-cosmos-go`). That app signs a gno transaction as a generic amino
document. It knows nothing of gno messages, and its field allowlist is what
forced the fee rendering in [pr6173](pr6173_ledger_compatible_sign_doc.md).

[ledger-app-gno](https://github.com/gnolang/ledger-app-gno) is a Ledger app
for gno.land. It parses `std.GetSignaturePayload` output itself and shows each
supported message field by field: `bank.MsgSend`, `vm.m_call`, `vm.m_addpkg`,
`vm.m_run`, the package-enabling messages and sessions. It speaks its own
protocol (`docs/APDU.md` in that repository):

| | Cosmos app | Gno.land app |
|---|---|---|
| CLA | `0x55` | `0xE0` |
| path | `m/44'/118'/a'/0/i` | the same, and nothing else |
| public key | uncompressed | compressed (33 bytes) and the `g1` address |
| signature | DER | `R ‖ S`, 64 bytes, S already low |
| signed bytes | the amino sign doc | the same, `std.GetSignaturePayload` |

Both sign the same payload over SHA-256 with secp256k1, so a key, its address
and its signatures are identical whichever app is used.

## Decision

Add a client for the Gno.land app, `GnoApp` in `tm2/pkg/crypto/internal/ledger`,
behind the existing `SECP256K1` device interface, and choose the app at
discovery:

1. `DiscoverDefault` connects to the first Ledger device and sends the Gno.land app's
   `GET_APP_NAME`. If the answer is `Gno`, it returns a `GnoApp`.
2. If the device answers with another name or with a status word, as the
   Cosmos app, other apps and the dashboard do, it closes the connection and
   falls back to `FindLedgerCosmosUserApp`, exactly as before.
3. A locked device (`0x5515`), a device still busy after a review (`0x6901`)
   and a failed USB transfer are returned as they are. Falling through to the
   Cosmos lookup would replace them with an error pointing at the wrong cause.

`GnoApp` adapts the Gno.land app to what callers already expect, so `crypto/ledger`,
the keybase and every `gnokey` command are unchanged:

- `GetPublicKeySECP256K1` returns the compressed key. `getPubKeyUnsafe` already
  parses it with `btcec.ParsePubKey`, which accepts either SEC1 form; the
  interface comment now says so.
- `GetAddressPubKeySECP256K1` checks that the address the device shows has the
  requested prefix. The app shows gno addresses only.
- `SignSECP256K1` streams the payload in 255-byte chunks after a path chunk,
  then wraps the returned `R ‖ S` as DER for `convertDERtoBER`. The round trip
  is lossless.
- Right after a review the device answers `0x6901` (COMMAND_NOT_ACCEPTED) until
  its status page closes. Such a command was not processed, so it is resent
  every 250 ms for up to 10 s.
- It reaches the device through `hidDevice`, a small HID transport using
  ledger-go's framing but writing one 64-byte report per call. ledger-go's own
  transport passes all the reports of an APDU to a single write. On macOS with
  a Nano S Plus, after such a write for a 255-byte payload chunk (five
  reports), the next write timed out and the device stopped receiving until
  unplugged; `gnokey sign` failed with `hidapi: unknown failure`. Writing one
  report per call, as Ledger's Python tooling does, signs normally. The Cosmos
  app keeps ledger-cosmos-go's transport.
- Status words become `*GnoStatusError` with a readable meaning (rejected on
  the device, Packages setting off, payload not accepted, ...). ledger-go
  reports only `ErrorMessage(sw)`'s text, so the code is recovered from that
  text; `hidDevice` reports status words the same way.

## Alternatives considered

- **A gnokey flag or key type per app.** Keys and signatures are the same with
  both apps, so the keybase has nothing to record. Detecting the open app keeps
  existing Ledger keys working with either.
- **Return `R ‖ S` from the device interface.** It would change the contract
  the Cosmos client satisfies, for no gain over a lossless DER wrap.
- **Detect the app with the BOLOS `GET_APP_AND_VERSION` command (`CLA 0xB0`).**
  Asking the Gno.land app directly also proves it speaks the protocol expected.

## Consequences

- Users with the Gno.land app open get gno-aware review on the device; users with
  the Cosmos app see no change.
- `zondax/ledger-go` becomes a direct dependency (it was already indirect).
- Tests use a fake device that follows the Gno.land app's protocol and session rules
  with a real secp256k1 key, and fake HID reports for the transport. The client
  was also run against the app itself, in Speculos and on a Nano S Plus:
  `gnokey add ledger` with the address approved on the device, then
  `gnokey sign` of a `bank.MsgSend` after the full review, which
  `gnokey verify` accepts.
