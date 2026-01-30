# CLA Enforcement

Package deployers must acknowledge the Contributor License Agreement (CLA) by
including a hash of the CLA content in `MsgAddPackage` transactions. The chain
validates this hash against governance-controlled parameters.

## Parameters

| Parameter | Description |
|-----------|-------------|
| `vm.Params.CLADocURL` | URL to CLA document (e.g., raw GitHub URL with commit hash) |
| `vm.Params.CLAHash` | Expected hash of CLA content (empty = enforcement disabled) |

Each chain can configure its own CLA via governance, enabling per-chain CLA requirements.

## Hash Format

First 16 hex characters (8 bytes) of SHA-256. This is a tradeoff: collision-resistant
enough for CLA versioning while keeping transaction size small.

```go
func ComputeCLAHash(content string) string {
    hash := sha256.Sum256([]byte(content))
    return fmt.Sprintf("%x", hash)[:16]
}
```

```bash
printf 'I agree to the CLA.\n' | sha256sum | cut -c1-16
# b6faa56f8eec79eb
```

## Using gnokey

Sign once per chain, deploy many times:

```bash
# Sign CLA for a chain (queries chain for document URL, displays content, prompts for "agree")
gnokey cla sign --remote https://rpc.gno.land:443

# Sign CLA from local file
gnokey cla sign --url /path/to/cla.txt --remote https://rpc.gno.land:443

# Check CLA status (all chains)
gnokey cla status

# Check CLA status for specific chain
gnokey cla status --remote https://rpc.gno.land:443

# Deploy packages (uses stored CLA hash for that remote automatically)
gnokey maketx addpkg -pkgpath gno.land/r/demo/hello -pkgdir ./hello --remote https://rpc.gno.land:443 mykey
```

CLA hashes are stored per-remote in `$GNOHOME/config.toml`:

```toml
[zones."https://rpc.gno.land:443"]
cla_hash = "b6faa56f8eec79eb"

[zones."https://rpc.test5.gno.land:443"]
cla_hash = "a3d74e2544d091e8"
```

If the CLA changes on-chain, run `gnokey cla sign` again to update.

## For Wallet Developers

Include `cla_hash` field in `MsgAddPackage`:

```json
{
  "@type": "/vm.MsgAddPackage",
  "creator": "g1...",
  "package": { ... },
  "cla_hash": "b6faa56f8eec79eb"
}
```

Query current enforcement state:
```bash
gnokey query params/vm:p:cla_doc_url --remote https://rpc.gno.land:443
gnokey query params/vm:p:cla_hash --remote https://rpc.gno.land:443
```

Empty `cla_hash` = disabled, non-empty = hash required.
