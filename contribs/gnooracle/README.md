# gnooracle

A small off-chain **package-approval oracle** for gno.land chains running with
the `inert` code-submission policy (see [PR #5888](https://github.com/gnolang/gno/pull/5888)).

Under the `inert` policy, anyone may submit a package with `MsgAddPackage`, but
it is stored **inert** — not typechecked, not executed, not importable. A
package only becomes active once an address in the chain's `PkgApprovers` param
sends `MsgEnablePackage`.

`gnooracle` automates that approver role:

1. **Watches** new blocks over RPC.
2. **Extracts** `MsgAddPackage` transactions from each block.
3. **Typechecks** the submitted package off-chain (same typechecker the chain
   uses, resolving stdlibs and `examples/` from disk).
4. If it passes, **broadcasts** a `MsgEnablePackage` signed by the approver key,
   activating the package on-chain.

> The oracle proposes, the chain enforces. The oracle is untrusted for
> correctness: the validator re-runs `TypeCheckMemPackage` at `MsgEnablePackage`
> time and rejects ill-typed code. The oracle only decides *which* pending
> packages get proposed for activation, and *when* — keeping the typechecker off
> the critical block-execution path.

## Usage

```sh
gnooracle \
  --remote http://127.0.0.1:26657 \
  --chain-id dev \
  --mnemonic "$GNOORACLE_MNEMONIC" \
  --gno-root /path/to/gno
```

The approver key's address (derived from the mnemonic, account 0 / index 0)
**must** be listed in the chain's vm `PkgApprovers` param, and `code_submission_policy`
must be `inert`, otherwise the `MsgEnablePackage` transactions are rejected.

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--remote` | `http://127.0.0.1:26657` | RPC address of the node to watch |
| `--chain-id` | *(required)* | Chain ID used to sign approval transactions |
| `--mnemonic` | `$GNOORACLE_MNEMONIC` | BIP39 mnemonic of the approver key |
| `--gno-root` | auto-detected | gno repo root, used to resolve stdlibs and examples for typechecking |
| `--gas-fee` | `1000000ugnot` | Gas fee for approval transactions |
| `--gas-wanted` | `20000000` | Gas wanted for approval transactions |
| `--poll-interval` | `1s` | How often to poll for new blocks |
| `--start-height` | `0` | Height to start watching from (0 = current tip) |

## Limitations

- **Dev-grade key handling**: the approver key is supplied as a raw mnemonic
  (flag or `$GNOORACLE_MNEMONIC`). For production, back the signer with an
  encrypted on-disk keybase instead.
- **Import resolution is disk-based**: packages that import other *on-chain-only*
  packages (not present under `examples/`) will fail the local typecheck and not
  be approved. Such packages remain inert until approved by a more capable
  approver. A future version can resolve imports via `vm/qfile` RPC queries.
- **No catch-up persistence**: `--start-height` lets you replay from a given
  height, but the oracle keeps no on-disk cursor between runs.
