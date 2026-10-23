# amino:"skip_zero", and a zero SessionAddr off the wire

## Status

Proposed, in PR #6399.

## Context

`std.Signature.SessionAddr` names the session key behind a signature and is
the zero address for every master-key signature. When account sessions added
it (#5307, merged 2026-04-21), its comment said "Amino binary and JSON both
skip the field when the address is zero ([20]byte{}), so master-signed txs
pay no wire-size overhead." That held for the generated encoder, which then
tested the Go value (`SessionAddr != [20]byte{}`).

Four days later #5590 made genproto2 judge a field empty by its
`MarshalAmino` repr, to match the reflect codec: the two had been writing
different bytes for zero addresses, a consensus hazard for fields such as
`Header.ProposerAddress` and `CommitSig.ValidatorAddress`. The zero
`crypto.Address` reprs as its bech32 string, which is not empty, so since then
amino binary writes field 3 of every master-key signature as
`0x1a 0x28 g1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqluuxe`, 42 bytes. JSON never
wrote it: `omitempty` compares the Go value. The comment stayed, and
`TestAminoSignatureSessionAddrRoundTrip` checks only that a zero
`SessionAddr` decodes as a master signature, which it does.

So every master-signed transaction a Go client builds (gnokey, gnoclient)
carries those 42 bytes per signature. For a single secp256k1 signature they
also take the signature's entry in the transaction from 126 bytes to 168, past
what a one-byte length prefix holds, so a plain bank send is 43 bytes longer:
430 gas at the default `tx_size_cost_per_byte` of 10, and 43 bytes in every
node's block store. tm2-js-client leaves the field out, so the same signed
transaction encoded in Go and in JavaScript had different bytes and different
hashes.

No existing tag could leave the field out. Amino's field tags were
`json:"-"` and `json:",omitempty"`, `binary:"fixed64|fixed32|varint"` and
`amino:"unsafe|write_empty|nil_elements|reserved"`. None skips a field by its
Go value; `write_empty` only forces writing.

## Decision

Add a field tag, `amino:"skip_zero"`: the field is left out of the binary and
JSON encodings whenever its Go value is the zero value of its type, as
`reflect.Value.IsZero` decides, even where its repr would encode to bytes. A
missing field decodes back to the zero value. Since Go 1.22 `IsZero` agrees
with `==` against the zero value, so a float `-0.0` counts as zero and decodes
as `0`. Decoders still accept an explicitly encoded zero value, so data
written before a field took the tag stays readable.

Registration refuses the tag next to `amino:"write_empty"`, which says the
opposite, and on any field whose missing value, `defaultValue`, is not the
zero value: `time.Time` (1970-01-01) and pointers to anything but a struct (a
pointer to the zero value). Everywhere else a field the tag leaves out decodes
to exactly what was left out.

It is implemented in every encoder:

- Reflect codec: `encodeReflectBinaryStruct` and `encodeReflectJSONStruct`
  skip a tagged zero field before any repr conversion;
  `decodeReflectJSONStruct` sets a missing tagged field to zero even next to
  `omitempty`, which alone leaves the old value. The binary decoder already
  sets a missing field to `defaultValue`.
- genproto2: `writeFieldMarshal`, `writeFieldSize` and
  `writeFieldMarshalInline` wrap a tagged field in a test of its Go value,
  `goo.F != (T{})` for a comparable named type or array and
  `!reflect.ValueOf(&goo.F).Elem().IsZero()` otherwise. Where amino already
  leaves the zero value out (a nil pointer, or a non-AminoMarshaler field
  that `zeroCheck` tests), the tag generates exactly the code it would without
  it.
- genproto (genproto1, deprecated for gno.land): `ToPBMessage` leaves the pb
  field unset when the Go value is zero, so proto3 leaves it out too;
  `FromPBMessage` reads an unset pb field as the zero Go value without
  converting it, as the reflect decoder never calls `UnmarshalAmino` for a
  missing field; and `Is*ReprEmpty` counts a zero tagged field as empty.

`std.Signature.SessionAddr` takes the tag. Regenerating every `pb3_gen.go`
changes only `tm2/pkg/std` (the guard in `Signature`'s `MarshalBinary2` and
`SizeBinary2`) and the amino test types.

`parseStructInfoWLocked` now re-panics with the cause, so a refused tag says
why instead of only "panic parsing struct".

## Alternatives considered

- Correct the comment and keep the bytes: every master-signed transaction
  keeps paying for them.
- Make `crypto.Address.MarshalAmino` return "" for the zero address: it
  changes every zero address in binary and JSON alike, in every type, so it
  needs an audit of everywhere a zero address is hashed, signed or stored, and
  a coordinated upgrade.
- Test the Go value again in genproto2 alone: the reflect codec would still
  write the field, and the two codecs would disagree again, the hazard #5590
  removed.
- Test the Go value for every AminoMarshaler field in both codecs: it changes
  every zero AminoMarshaler field of every registered type, consensus types
  included, where one opt-in tag changes one field.
- Give `SessionAddr` its own type with an empty repr when zero, or make it a
  `*crypto.Address`: both change the Go API of `std.Signature` for gnokey,
  gnoclient, the ante handler and users outside the repo, such as the x402
  facilitator.

## Consequences

- Wire: a master-key signature built by a Go client is 42 bytes shorter, and
  a single-signature secp256k1 transaction 43. Go now encodes such a
  transaction exactly as tm2-js-client does. Session-key signatures encode as
  before. A placeholder signature, all zero as gnoclient fills one in before
  signing, is now an empty list element of 2 bytes and still counts.
- Gas: 430 less for a single-signature send at the default params. The fee is
  what the sender set, so it changes only where a client derives it from
  simulated gas.
- Nodes: `CheckTx`, `DeliverTx` and simulation run the sender's own bytes,
  both encodings decode to the same transaction, the sign bytes never covered
  the signature's encoding, and no signature is stored in state. Nodes with
  and without this change agree on every transaction, and wallets that still
  send the old encoding keep working.
- Genesis and hardfork replay are unaffected. `BaseApp.Deliver`, which
  InitChain uses for each of those transactions, re-encodes the transaction
  but runs it with no transaction bytes in its context, so the ante handler
  charges them no size gas whatever the encoding.
- Hashes: `gnogenesis txs remove` and its de-duplication hash a transaction
  by re-encoding it from JSON. They now reproduce the on-chain hash of
  transactions built by JavaScript clients and by Go clients from now on, and
  no longer of those Go clients built between #5590 and this change.
  `tx-archive` re-encodes transactions on restore, so restored ones of the
  latter kind get new hashes.
- `amino:"skip_zero"` is opt-in. Adding it to an existing field changes the
  bytes written for that field's zero value, binary and JSON, which is a
  protocol change wherever they are hashed or signed; the amino README says
  so. `SessionAddr` is in neither the sign bytes nor its JSON (it was already
  omitted there).

`tm2/pkg/amino/tests/skip_zero_test.go` pins the tag's behaviour across the
codecs, `TestSkipZeroGuard` and `TestGenerateProtoBindings_SkipZero` the
generators' output, `TestAminoSignatureZeroSessionAddrWire` the new and old
`SessionAddr` encodings, and `TestDeliverTxZeroSessionAddr` that a chain takes
both.
