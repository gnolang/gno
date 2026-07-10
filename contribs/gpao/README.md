# gpao

**gpao** (gno package-approver oracle) is a small off-chain approval daemon for
gno.land chains running with the `inert` code-submission policy (see
[PR #5888](https://github.com/gnolang/gno/pull/5888)).

Under the `inert` policy, anyone may submit a package with `MsgAddPackage`, but
it is stored **inert** — not typechecked, not executed, not importable. A
package only becomes active once an address in the chain's `PkgApprovers` param
sends `MsgEnablePackage`.

`gpao` automates that approver role:

1. **Watches** new blocks over RPC.
2. **Extracts** `MsgAddPackage` transactions from each block.
3. **Typechecks** the submitted package off-chain (same typechecker the chain
   uses). Imports resolve from the local disk store (stdlibs + `examples/`)
   first, falling back to `vm/qfile` RPC queries against the watched node for
   on-chain-only packages.
4. If it passes, **broadcasts** a `MsgEnablePackage` signed by the approver key,
   activating the package on-chain.

> The oracle proposes, the chain enforces. gpao is untrusted for correctness:
> the validator re-runs `TypeCheckMemPackage` at `MsgEnablePackage` time and
> rejects ill-typed code. gpao only decides *which* pending packages get
> proposed for activation, and *when* — keeping the typechecker off the critical
> block-execution path.

## Usage

The approver key lives in a local [gnokey](../../gno.land/cmd/gnokey) keystore.
Create one, fund it, add its address to the chain's `PkgApprovers` param, then:

```sh
gpao \
  --remote http://127.0.0.1:26657 \
  --chain-id dev \
  --home ~/.gnokey \
  --key approver
```

gpao unlocks the key at startup: it reads the password from `$GPAO_PASSWORD` if
set (for unattended/service deployments), otherwise prompts once interactively.

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--remote` | `http://127.0.0.1:26657` | RPC address of the node to watch |
| `--chain-id` | *(required)* | Chain ID used to sign approval transactions |
| `--home` | gnokey home (`$GNOHOME`) | Keystore directory holding the approver key |
| `--key` | *(required)* | Name or bech32 address of the approver key |
| `--gno-root` | auto-detected | gno repo root, used to resolve stdlibs and examples for typechecking |
| `--gas-fee` | `1000000ugnot` | Gas fee for approval transactions |
| `--gas-wanted` | `20000000` | Gas wanted for approval transactions |
| `--poll-interval` | `1s` | How often to poll for new blocks |
| `--start-height` | `0` | Height to start watching from (0 = current tip) |

The key's address **must** be listed in the chain's vm `PkgApprovers` param, and
`code_submission_policy` must be `inert`, otherwise the `MsgEnablePackage`
transactions are rejected.

### Signing options

- **Local gnokey keystore (default, recommended)** — the encrypted key stays on
  disk and is unlocked at startup. This is the same keystore `gnokey` uses.
- **Mnemonic (dev only)** — set `$GPAO_MNEMONIC` to sign from a raw mnemonic
  without a keystore. Convenient for local devnets; not for production.
- **tmkms / gnokms are NOT supported** — those are *consensus* key managers that
  sign block votes over the privval protocol. gpao signs application
  *transactions* (`MsgEnablePackage`), which they cannot do. Use the gnokey
  keystore (or, in future, an HSM/KMS-backed keystore that can sign txs).

## Limitations

- **RPC import cache is per-run**: on-chain packages fetched via `vm/qfile` are
  cached for the process lifetime, so a dependency updated mid-run is not
  re-fetched. This only affects the oracle's *proposal*; the chain re-typechecks
  against current state at enable time.
- **No catch-up persistence**: `--start-height` lets you replay from a given
  height, but gpao keeps no on-disk cursor between runs.
