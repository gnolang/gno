# Charge chain.emit per event byte, typ included

## Context

`chain.emit` was priced by its gas-table row on the number of `attrs` elements
alone: 362 gas plus 39.3 gas per element. A 4096-byte attribute cost the same as
a 1-byte one, and the 4096-byte `typ` was free.

Every byte an event carries outlives the call. The `EventLogger` keeps it until
the transaction ends, the block keeps it in `ABCIResponses` until
`SaveABCIResponses`, and in between it is amino-encoded twice: once per result
for `LastResultsHash`, once more for the state DB write. Nothing caps the total
per transaction or per block.

Measured on this path (`ResultsHash` then `SaveABCIResponses` on pebbledb,
134 MB of event payload, AMD Ryzen 7 7840HS): 3.7 to 4.4 ns per payload byte, a
retained heap of 1.0 byte per payload byte, and a peak heap in use of about 6.7
bytes per payload byte while the two encodings run.

Measured gas for one more `chain.Emit` call in a realm loop, before this change:

| call | gas | payload | per 3B-gas block | peak heap, extrapolated |
|---|---:|---:|---:|---:|
| `Emit("t", "k", v)`, `len(v)` 4096 | 3,557 | 4,098 B | 3.5 GB | 23 GB |
| `Emit(typ)`, `len(typ)` 4096 | 3,296 | 4,096 B | 3.7 GB | 25 GB |

`gnoland-1` runs with a 3,000,000,000 block gas limit, so one block of such
calls exhausts the memory of any validator on `ApplyBlock`.

## Decision

`X_emit` charges the machine's gas meter 14 gas per byte of the amino encoding
of the event it builds: `typ`, every key and value, the emitting package path
and the framing around them. 14 is `store.DefaultGasConfig().WriteCostPerByte`,
the price the chain already charges for writing one byte to its store, and every
event byte is written to the state DB. The measured 3.7 to 4.4 ns per byte of
encoding and hashing fits under it.

The charge runs after the length checks, so it only ever prices bytes the event
keeps. Billing the encoding rather than the raw lengths prices what the state DB
stores: an empty attribute pair has no bytes of its own and still pays for the 2
bytes of framing it adds. The gas-table row keeps the CPU fit of the call, and
its comment names this charge.

`eventSize` computes the encoding's length from the field lengths instead of
encoding the event: encoding it only to measure it cost as much CPU as the
charge bills on 64 empty pairs. `X_emit` also copies `typ`, keys and values, so
an emitted substring does not keep its whole source string alive until the
block ends, which would retain far more than the 1.0 byte per payload byte above.

After the change, the same calls cost 61,615 and 61,228 gas from
`TestEmitChargesPerByte`'s realm, which caps one block at about 200 MB of event
payload and a peak heap near 1.4 GB.

## Alternatives considered

- A per-byte slope in the gas-table row, `Slope2` on `SizeSliceTotalBytes`
  over `attrs`. The payload moves to `typ`, which still costs 3,296 gas per
  4096 bytes. Covering `typ` too needs a third pre-call slot or a new size kind
  reading the `(typ, attrs)` pair, for one caller, and the table's rows are CPU
  fits that `gen_native_table.py` regenerates: a hand-set storage price there
  is dropped on the next run.
- A charge in tm2's `EventLogger.EmitEvent`. The logger holds no gas meter,
  and bank transfers and storage deposits emit through it too, so their gas
  would move with a change aimed at realm events.
- A rate set from memory rather than from the store price, which is what the
  stored bytes cost.
- A per-transaction or per-block cap on event bytes. The per-byte price bounds
  the byte-driven growth by itself; a cap belongs to a separate change if one
  is wanted.

## Consequences

- An honest event pays 14 gas per encoded byte on top of its previous price:
  a two-pair event with 100 bytes of type and attributes, from a 20-byte
  package path, encodes to 150 bytes and pays 2,100 gas.
- The CPU the block path spends per event and per attribute stays priced by
  the gas-table row's CPU fit.
