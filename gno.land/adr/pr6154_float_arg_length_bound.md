# ADR-6154: Bound the length of a float call argument

## Context

`convertFloat` in `gno.land/pkg/sdk/vm/convert.go` turns a `MsgCall` string
argument into a `float32` or `float64` through `apd.NewFromString`. apd converts
the whole mantissa into a `big.Int` with `big.Int.SetString` before it looks at
the exponent, and prints it back through `Decimal.String()` before
`strconv.ParseFloat` sees it. Both conversions cost time that grows with the
square of the digit count.

`VMKeeper.Call` converts the arguments before it installs the machine's
recover, and the conversion never touches the gas meter or the allocator. The
only charge covering the argument bytes is the ante handler's
`TxSizeCostPerByte`, 10 gas per byte, which is linear. A one-megabyte float
argument, the size a single transaction may carry under `MaxBlockTxBytes`,
therefore buys seconds of block-execution time for about ten million gas out of
a three-billion gas block.

Measured on this branch's parent, one argument through `convertArgToGno`:

| argument bytes | float64 | string |
| --- | --- | --- |
| 128 | 5.9 µs | 63 ns |
| 1,024 | 61 µs | 54 ns |
| 4,096 | 290 µs | 49 ns |
| 16,384 | 1.9 ms | 50 ns |
| 100,000 | 39 ms | 48 ns |

## Decision

Refuse a float argument longer than 1024 bytes, in `convertFloat`, before apd
sees it. The check is a length comparison, so the quadratic conversion never
starts.

The bound is where the quadratic term stops being noise. Per byte the cost is
46 ns at 128 bytes and 60 ns at 1024, against 116 ns at 16 KB and 390 ns at
100 KB. A block's two megabytes of transaction data therefore cost about 0.12
seconds of parsing at the bound, against seconds without it, and a shorter
bound buys little: the per-byte cost below 1024 bytes is flat.

Every `float64` value is reachable under the bound. Seventeen significant
digits and an exponent name any of them, and the longest exact decimal
expansion, the smallest subnormal at 1074 fractional digits, is a spelling of a
value that `5e-324` also names.

## Alternatives considered

Charge gas proportional to the parse instead of bounding the length. Metering
is the answer where the value is arbitrary-precision and a bound would refuse
a program the language accepts, which is the case for the numeric literals the
GnoVM parses out of source. A call argument is not that: it is on its way into
a 64-bit float, and no caller loses a value to the bound. Metering here would
also mean giving `convertArgToGno` a gas meter it does not have.

Bound every argument rather than the float ones. The other conversions are
linear: `strconv.ParseInt` and `ParseUint` stop at the first overflow, base64
decoding is proportional, and a string argument is copied.

That linearity is about the parse, not the whole branch. Two of them still do
work an argument's own length does not pay for, and one is fixed here; see
below.

## Also decided: pin a `[N]byte` argument to its canonical encoded length

`convertArgToGno`'s `*gno.ArrayType` case decoded the whole base64 argument
and only then compared the result against the declared array length — the same
work-before-validation shape as the float case, one branch below it.
`DecodeString` allocates `DecodedLen(len(arg))` up front, so a one-megabyte
argument to a `[32]byte` parameter allocated about 750 KB, decoded all of it,
and was then rejected for length. Refusing on `len(arg)` first costs nothing.

The check has to be an equality against `EncodedLen(bt.Len)` rather than an
upper bound, because base64's decoder ignores `\r` and `\n`: a payload may
carry arbitrarily many of them, so `len(arg)` bounds nothing on its own. A
one-megabyte argument consisting of newlines and a valid 32-byte encoding was
accepted before this change.

Requiring the canonical length therefore also removes that padding as a source
of malleability, which the header comment on these conversions forbids
("very important that there is no malleability"): it previously spelled one
array value in unboundedly many ways. It does not remove every such spelling —
the decoder is also lenient about the unused trailing bits of the final
quantum, which is untouched here.

The equality does not subsume the decoded-length check, which stays. Padding
makes `EncodedLen` constant across three inputs — 1, 2 and 3 bytes all encode
to 4 characters — so an argument of the right encoded length can still decode
to `bt.Len`±2 bytes.

The `*gno.SliceType` (`[]byte`) case is left alone: it has no declared length
to pin an argument to, and its decode is proportional to an argument that
already pays for its own bytes.

## Consequences

A `MsgCall` carrying a float argument over 1024 bytes now fails where it
previously succeeded, so the change is consensus-affecting and needs a
coordinated upgrade. No value becomes unreachable, only the spellings of it
past the bound.

The panic reaches `BaseApp.runTx`'s recover rather than the keeper's, as every
other argument-conversion panic already does, and the transaction fails with
the argument's own length in the message.

The `[N]byte` change is consensus-affecting on the same terms. A `MsgCall`
whose `[N]byte` argument was padded with newlines now fails where it
previously succeeded. No value becomes unreachable — the canonical encoding of
every `[N]byte` still converts — and no test or fixture in the tree passes a
`[N]byte` call argument, so nothing in the repo changes behaviour. Note that
the `base64` CLI wraps its output at 76 columns by default, so a caller who
piped it unwrapped (`base64 -w0`) is unaffected while one who did not now has
to be.

Neither bound is gated on block height, so a hardfork replaying its own
history through `deliverGenesisTx` will refuse a historical `MsgCall` that
carried an over-long float argument or a newline-padded `[N]byte` one, and
`PanicOnFailingTxResultHandler` — the default — turns that into a panic inside
`InitChain` rather than a skipped transaction. Reproduced: a genesis tx
carrying a 1025-byte float argument panics the node at boot, while 1024 bytes
replays. `checkCodePolicy` carries the precedent for exempting replay
(`auth.IsGenesisReplay`), but `convertArgToGno` has no `sdk.Context` to consult,
so wiring one in is a larger change than either bound. Recorded here as a known
consequence for whoever schedules the upgrade; no chain in the tree is known to
carry such a transaction.
