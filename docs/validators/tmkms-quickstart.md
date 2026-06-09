# gnoland + tmkms: end-to-end setup

A hands-on, copy-paste walkthrough that stands up a single-host gnoland
validator whose consensus key is held by [tmkms] instead of sitting on
the node. It takes you from an empty directory to a node producing
blocks that are signed remotely by tmkms.

This is the practical companion to [tmkms.md](tmkms.md) — read that for
the architecture, the security model, and the production checklist. This
page is the "just make it work once" path.

[tmkms]: https://github.com/iqlusioninc/tmkms

> **Single host, softsign, for learning.** This walkthrough runs tmkms
> and gnoland on the same machine with tmkms's file-based `softsign`
> backend. That is fine for understanding the wiring and for testnets,
> but it is **not** a production layout — the consensus key still touches
> the validator host's disk. For real stake, move tmkms to its own host
> with an HSM/Ledger/cloud-KMS backend. See
> [tmkms.md § When to use this mode](tmkms.md#when-to-use-this-mode).

## How the pieces connect

```
  gnoland (validator)                 tmkms (signer)
  ───────────────────                 ──────────────
  listens on :26659  ◄────dials in──── [[validator]] addr = tcp://<peer_id>@…:26659
  allowed_kms_pubkeys ─ verifies ────► kms-identity.key   (SecretConnection identity)
  node_id (peer ID)  ◄─ verifies ───── <peer_id> pinned in addr
  genesis validator  = consensus pubkey = consensus.key   (the key that signs votes)
  chain_id           ══════ must match ══════ [[chain]].id / [[validator]].chain_id
```

The connection is **mutually** authenticated: gnoland verifies tmkms via
`allowed_kms_pubkeys`, and tmkms verifies gnoland via the `<peer_id>`
pinned in its `addr`. Both directions are mandatory here.

Three distinct ed25519 keys are in play — don't mix them up:

| Key | Lives in | Role |
|---|---|---|
| **consensus key** | tmkms `consensus.key` (softsign) | signs votes/proposals; its pubkey is the validator's identity in genesis |
| **kms-identity key** | tmkms `kms-identity.key` | tmkms's SecretConnection identity; its pubkey goes in gnoland's `allowed_kms_pubkeys` |
| **node key** | gnoland `node_key.json` | gnoland's SecretConnection identity; its peer ID (hex) is pinned in tmkms's `addr` |

The defining property of this mode: **tmkms dials gnoland**, not the
other way around. gnoland listens; tmkms connects in.

## 0. Prerequisites

- **Go** (to build `gnoland`).
- **Rust / cargo** (to install tmkms).
- **tmkms 0.15.0 built with the `softsign` feature.** The default
  release does *not* include softsign — install it explicitly:

  ```sh
  cargo install tmkms --version 0.15.0 --features softsign --locked
  tmkms version   # → 0.15.0
  ```

- **`gnoland`** built from this repo:

  ```sh
  # from the repo root
  go build -o ./build/gnoland ./gno.land/cmd/gnoland
  ```

Pick a working directory and a chain ID you'll reuse everywhere:

```sh
export WORK=~/gno-tmkms-lab
export CHAIN_ID=gno-tmkms-test
export GNOLAND=$(pwd)/build/gnoland   # absolute path to the binary you just built

mkdir -p "$WORK"/gnoland-data/secrets "$WORK"/tmkms/secrets
cd "$WORK"
```

> The `chain_id` must be **identical** in three places: the gnoland
> genesis (`-chainid`), the `tmkms_listener.chain_id` in `config.toml`,
> and the `[[chain]].id` / `[[validator]].chain_id` in `tmkms.toml`. If
> they drift, tmkms refuses to sign or the signatures won't verify.

## 1. Generate the gnoland node secrets

This creates the consensus key, the node's p2p key, and the sign-state
file:

```sh
"$GNOLAND" secrets init -data-dir ./gnoland-data/secrets
```

You now have `gnoland-data/secrets/priv_validator_key.json` (the
consensus key), `node_key.json` (p2p identity), and
`priv_validator_state.json`.

> In tmkms mode gnoland does **not** use `priv_validator_key.json` to
> sign at runtime — tmkms does. But gnoland *does* read it once, at
> genesis-generation time, to learn the validator's pubkey. So the
> consensus key in this file and the key you load into tmkms (next step)
> **must be the same key**. We achieve that by exporting this one into
> tmkms rather than generating a second one.

Note the validator address — you'll confirm it matches the genesis later:

```sh
"$GNOLAND" secrets get -data-dir ./gnoland-data/secrets validator_key
```

You also need the node's **peer ID** in hex — tmkms uses it to verify it
is talking to *your* validator and not an impostor on the listen port
(step 4 pins it). gnoland's `secrets get node_id` reports this identity
in bech32 (`g1…`); tmkms wants the same 20 bytes in hex. This tiny
stdlib-only Go helper reads `node_key.json` and prints the hex form:

```sh
cat > nodeid-hex.go <<'GO'
// Prints a gnoland node's peer ID in the hex form tmkms expects.
package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	raw, err := os.ReadFile(os.Args[1]) // path to node_key.json
	if err != nil {
		panic(err)
	}
	var nk struct {
		PrivKey string `json:"priv_key"`
	}
	if err := json.Unmarshal(raw, &nk); err != nil {
		panic(err)
	}
	priv, err := base64.StdEncoding.DecodeString(nk.PrivKey)
	if err != nil {
		panic(err)
	}
	pub := priv[32:]           // ed25519: priv = seed(32) ‖ pub(32)
	addr := sha256.Sum256(pub) // node address = SHA256(pubkey)[:20]
	fmt.Println(hex.EncodeToString(addr[:20]))
}
GO

export VALIDATOR_PEER_ID=$(go run ./nodeid-hex.go ./gnoland-data/secrets/node_key.json)
echo "$VALIDATOR_PEER_ID"   # e.g. 243cef06…dcac — pinned into tmkms.toml in step 4
```

## 2. Export the consensus key into tmkms's softsign format

gnoland stores the ed25519 private key as base64 of the 64-byte
`seed‖pubkey`. tmkms softsign wants base64 of just the 32-byte **seed**.
Slice off the first 32 bytes and re-encode:

```sh
python3 - <<'PY' > ./tmkms/secrets/consensus.key
import json, base64
v = json.load(open("gnoland-data/secrets/priv_validator_key.json"))["priv_key"]["value"]
print(base64.b64encode(base64.b64decode(v)[:32]).decode(), end="")
PY
chmod 600 ./tmkms/secrets/consensus.key
```

(No crypto involved — it's a byte reslice — so plain `python3` is enough.)

## 3. Generate tmkms's SecretConnection identity key

tmkms needs its own identity key to authenticate the inbound connection.
gnoland will only accept the connection if this key's **public** half is
in `allowed_kms_pubkeys`. Generate the key and print its hex pubkey with
this tiny Go helper:

```sh
cat > tmkms-identity-keygen.go <<'GO'
package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
)

func main() {
	seed := make([]byte, ed25519.SeedSize)
	if _, err := rand.Read(seed); err != nil {
		panic(err)
	}
	pub := ed25519.NewKeyFromSeed(seed).Public().(ed25519.PublicKey)
	// tmkms softsign connection key = base64 of the 32-byte seed.
	if err := os.WriteFile(os.Args[1], []byte(base64.StdEncoding.EncodeToString(seed)), 0o600); err != nil {
		panic(err)
	}
	// Hex pubkey for gnoland's allowed_kms_pubkeys.
	fmt.Println("ed25519:" + hex.EncodeToString(pub))
}
GO

ALLOW=$(go run ./tmkms-identity-keygen.go ./tmkms/secrets/kms-identity.key)
echo "$ALLOW"   # e.g. ed25519:4b6efade…b18a — copy this for step 5
```

## 4. Write `tmkms.toml`

This is the format verified by the repo's real-tmkms integration test
(`tm2/pkg/bft/privval/upstream/tmkms_integration_test.go`). Note
`protocol_version = "v0.34"` — gnoland only speaks that dialect and
refuses to start otherwise.

```sh
cat > ./tmkms/tmkms.toml <<TOML
[[chain]]
id = "${CHAIN_ID}"
key_format = { type = "hex" }
state_file = "$(pwd)/tmkms/secrets/consensus_state.json"

[[providers.softsign]]
chain_ids = ["${CHAIN_ID}"]
key_type = "consensus"
key_format = "base64"
path = "$(pwd)/tmkms/secrets/consensus.key"

[[validator]]
chain_id = "${CHAIN_ID}"
addr = "tcp://${VALIDATOR_PEER_ID}@127.0.0.1:26659"
secret_key = "$(pwd)/tmkms/secrets/kms-identity.key"
protocol_version = "v0.34"
reconnect = true
TOML
```

- `addr` is where gnoland is listening (step 6), with the
  **validator peer ID from step 1 pinned in front** (`<peer_id>@host:port`,
  Tendermint-style). This is **required**, not cosmetic — it is the
  mirror of gnoland's `allowed_kms_pubkeys`, and it makes the
  authentication mutual:

  | Direction | Authenticated by |
  |---|---|
  | gnoland verifies tmkms | `allowed_kms_pubkeys` (step 5) |
  | tmkms verifies gnoland | the peer ID pinned in `addr` (here) |

  Omit the `<peer_id>@` prefix and tmkms will sign for *whoever* answers
  on the port and log `unverified validator peer ID!` on every
  connection. On a single host `127.0.0.1` is fine; across hosts use the
  validator's reachable IP (the peer ID stays the same).
- `state_file` is tmkms's `consensus.json` — the authoritative
  double-sign (HRS) gate. tmkms creates it on first run. **Never restore
  a stale copy of it.**
- `reconnect = true` makes tmkms keep retrying the dial, so you can
  start it before or after gnoland.

> tmkms 0.15.0 prints `deprecated protocol_version v0.34 (update to
> v0.38)! Will be a hard error in next release` on startup. That's
> expected: gnoland speaks **only** v0.34 today (see
> [tmkms.md § Protocol version pin](tmkms.md#protocol-version-pin)), so
> stay on a tmkms release that still supports v0.34 — 0.15.0 does.

## 5. Configure gnoland for tmkms mode

Create a default `config.toml`, then point the `tmkms_listener` block at
tmkms.

```sh
CFG=./gnoland-data/config/config.toml
"$GNOLAND" config init -config-path "$CFG"
```

> **Ordering gotcha.** `config set` validates the whole `tmkms_listener`
> block on every write, and a non-empty `listen_addr` is what *enables*
> validation of the other fields. So set `listen_addr` **last** — if you
> set it first, the write is rejected because `chain_id` is still empty,
> and it silently stays unset.

```sh
"$GNOLAND" config set -config-path "$CFG" consensus.priv_validator.tmkms_listener.chain_id "$CHAIN_ID"
"$GNOLAND" config set -config-path "$CFG" consensus.priv_validator.tmkms_listener.protocol_version "v0.34"
"$GNOLAND" config set -config-path "$CFG" consensus.priv_validator.tmkms_listener.allowed_kms_pubkeys "$ALLOW"
# listen_addr LAST — this enables the mode:
"$GNOLAND" config set -config-path "$CFG" consensus.priv_validator.tmkms_listener.listen_addr "tcp://0.0.0.0:26659"
```

Verify it stuck:

```sh
"$GNOLAND" config get -config-path "$CFG" consensus.priv_validator.tmkms_listener
```

`listen_addr` should be `tcp://0.0.0.0:26659` and not empty.

## 6. Start tmkms, then gnoland

**Order matters.** gnoland blocks at startup for up to
`wait_for_connection_timeout` (default 60s) waiting for tmkms to dial in,
and it needs that connection during its own startup. Start tmkms first.

Terminal 1 — tmkms:

```sh
tmkms start -c ./tmkms/tmkms.toml
```

You should see it load the softsign provider and (once gnoland is up)
complete the SecretConnection handshake. With `reconnect = true` it logs
connection-refused retries until gnoland is listening — that's expected.

Terminal 2 — gnoland:

```sh
"$GNOLAND" start \
  -data-dir ./gnoland-data \
  -genesis ./gnoland-data/genesis.json \
  -chainid "$CHAIN_ID" \
  -lazy \
  -skip-genesis-sig-verification
```

- `-genesis ./gnoland-data/genesis.json` keeps the generated genesis
  alongside the rest of the node state. Without it, `genesis.json` is
  written to your **current working directory**.
- `-skip-genesis-sig-verification` is needed because the default
  `-lazy` genesis includes example deploy txs signed by test accounts;
  without this flag the node panics replaying them at boot. It has
  nothing to do with tmkms — it's the standard dev-genesis flag (the
  same one in the gnoland README quick-start).

What `-lazy` does here, in order:

1. Keeps your existing `config.toml` and `secrets/` (it only generates
   files that are missing — it will **not** clobber your tmkms config).
2. Generates `genesis.json`, reading the **local** consensus pubkey from
   `priv_validator_key.json` to register this node as the genesis
   validator. (This is why the exported key in step 2 must match.)
3. Boots the node. The privval stack now uses the tmkms listener: it
   waits for tmkms's dial-in, fetches the pubkey over the link, and from
   then on every vote/proposal is signed by tmkms.

## 7. Verify it's really tmkms signing

Within a few seconds you should see the node advancing through heights.

- **tmkms log**: `connected to validator successfully`, then ongoing
  sign activity. That's the signer doing the work.
- **gnoland log**: `Signed and pushed vote …`, `Finalizing commit of
  block`, and `Committed state … height=N` climbing. If signing were
  broken the node would stall at height 1 instead of advancing.

  ```sh
  grep "Committed state" ./gnoland.log | tail
  ```

- **Identity check** — the validator address gnoland signs with must be
  the address from step 1:

  ```sh
  "$GNOLAND" secrets get -data-dir ./gnoland-data/secrets validator_key
  # the "address" must match the `validator address` in the
  # "Signed and pushed vote" log lines, and the validator entry in
  # ./gnoland-data/genesis.json
  ```

- **Mutual auth holds** — tmkms logs clean
  `signed Proposal/Prevote/Precommit …` lines and **no**
  `unverified validator peer ID!` warning. If you do see that warning,
  the peer ID pinned in `tmkms.toml`'s `addr` (step 4) is missing or
  wrong — tmkms prints the value it expected in the parentheses, so
  compare it against `$VALIDATOR_PEER_ID` from step 1.

- **Kill tmkms** and watch gnoland stop committing blocks (the height
  stops climbing; the log shows the signer connection dropping). Restart
  tmkms and it resumes. That's the clearest proof the key lives in
  tmkms, not gnoland.

> **Benign log noise.** Even when signing works, gnoland periodically
> logs `SignerListener: accept failed … i/o timeout` and `already
> connected, dropping listen request`. The endpoint holds one live
> signer connection and these are the idle re-accept attempts timing
> out; they don't interrupt signing. Watch the committed height, not
> these lines.

## Troubleshooting

| Symptom | Cause | Fix |
|---|---|---|
| tmkms: `unknown field 'softsign', expected 'ledgertm'` | tmkms built without the softsign feature | Reinstall: `cargo install tmkms --version 0.15.0 --features softsign --locked` |
| gnoland: `tmkms_listener.allowed_kms_pubkeys must not be empty` | allowlist not set, or `listen_addr` was set before the other fields | Set the four fields with `listen_addr` **last** (step 5) |
| gnoland exits after ~60s waiting for connection | tmkms isn't running, can't reach `addr`, or its pubkey isn't in the allowlist | Start tmkms first; check `addr` matches `listen_addr`; confirm the hex in `allowed_kms_pubkeys` is the one printed in step 3 |
| gnoland: `protocol_version` rejected at startup | a value other than `v0.34` | Use `v0.34` on both sides |
| gnoland panics at boot: `PubKey does not match Signer address … InvalidPubKeyError` | default `-lazy` genesis has example txs signed by test accounts | Add `-skip-genesis-sig-verification` (step 6) — unrelated to tmkms |
| Node never advances past height 1, no signing in tmkms log | `chain_id` mismatch between genesis, `config.toml`, and `tmkms.toml` | Make all three identical and restart |
| Signatures produced but rejected by the chain | consensus key in tmkms ≠ genesis validator key | Re-export in step 2 from the *same* `priv_validator_key.json` used for genesis |
| gnoland logs `accept failed: i/o timeout` / `already connected, dropping listen request` while still committing blocks | benign idle re-accept attempts on the held connection | Ignore — confirm the committed height is climbing |
| tmkms: `unverified validator peer ID! (<hex>)` | the peer ID prefix is missing from `tmkms.toml`'s `addr`, so tmkms signs for whoever answers on the port | Pin it: `addr = "tcp://<hex>@host:port"` using `$VALIDATOR_PEER_ID` from step 1 (step 4) |
| tmkms (Ledger): startup `SigningError` / can't get pubkey | device locked, asleep, or not on the Tendermint Validator app | Unlock the Ledger and open the **Tendermint Validator** app, then restart tmkms (L0) |
| gnoland (Ledger): boots but `This node is NOT a validator`, height stuck at 0 | the `gpub`/address in genesis doesn't match the device key | Re-read the hex from tmkms's `added consensus Ed25519 key` log (L2), re-run `pkconv.go` (L3), rebuild genesis (L4) |

## Variant: a Ledger hardware signer instead of softsign

softsign keeps the consensus key in a file — fine for a lab, but the
key still touches disk. A Ledger holds the key **on the device** and
never exports it. That single fact flips the bootstrap around:

> With softsign you generated the key in gnoland and *pushed it into*
> the signer (step 2). With a Ledger you can't — the private key never
> leaves the device. So you go the other way: read the **public** key
> *out of* the Ledger and seed the genesis validator from it.

Everything about the connection (listener, allowlist, peer-ID pin,
`protocol_version`, `chain_id`) is identical. Only the key custody and
the genesis-bootstrap change.

> **Scope note.** The gno-side of this flow — converting the device
> pubkey, seeding genesis with it, and signing through the tmkms
> listener — is verified end-to-end. The physical-device half (the
> `ledgertm` provider talking to a real Ledger) is described from the
> tmkms source and the Ledger app; it needs the hardware to exercise.
> A Ledger is also a single signer with no failover — for real
> high-availability use Horcrux or an HSM (see [tmkms.md](tmkms.md)).

### L0. Prerequisites

- tmkms built with the **`ledgertm`** provider. It's in the default
  build (`cargo install tmkms` with no `--features` gives you exactly
  `ledgertm`); to keep softsign too, build
  `--features ledgertm,softsign`.
- A Ledger with the **"Tendermint Validator"** app installed (the
  dedicated ed25519 consensus-signing app — *not* the regular Cosmos
  app), plugged in, unlocked, with that app **open**. tmkms can't reach
  a locked device or a different app.

### L1. Point tmkms at the Ledger

Same `tmkms.toml` as step 4, but replace the `[[providers.softsign]]`
block with a `[[providers.ledgertm]]` block — no key files, no paths:

```toml
[[chain]]
id = "${CHAIN_ID}"
key_format = { type = "hex" }       # makes tmkms log the pubkey in hex (L2)
state_file = "/abs/path/tmkms/secrets/consensus_state.json"

[[providers.ledgertm]]
chain_ids = ["${CHAIN_ID}"]

[[validator]]
chain_id = "${CHAIN_ID}"
addr = "tcp://${VALIDATOR_PEER_ID}@127.0.0.1:26659"   # peer-ID pin from step 1
secret_key = "/abs/path/tmkms/secrets/kms-identity.key"
protocol_version = "v0.34"
reconnect = true
```

Only one `[[providers.ledgertm]]` is allowed, and it always signs the
**consensus** key.

### L2. Read the consensus pubkey out of the Ledger

Start tmkms alone. On connect it logs the device's consensus pubkey —
in hex, because of `key_format = { type = "hex" }`:

```sh
tmkms start -c ./tmkms/tmkms.toml
# [keyring:ledgertm] added consensus Ed25519 key: 2C854661478AA1CDC954D11ABA6ABB6DBF469572564C24C61ABFC0622A04D350
```

Copy that hex string — it's your validator's consensus public key.

### L3. Convert the hex pubkey to gno format

gnoland's genesis wants the pubkey as a `gpub1…` bech32 and the matching
`g1…` address. This stdlib-plus-gno-crypto helper converts the hex (run
it **from the gno repo root** so the imports resolve):

```sh
cat > pkconv.go <<'GO'
// Converts a raw ed25519 consensus pubkey (hex or base64) to gno's
// gpub1… / g1… forms. Run from the gno repo root: go run ./pkconv.go <hex>
package main

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
)

func main() {
	in := strings.TrimSpace(os.Args[1])
	bz, err := hex.DecodeString(in)
	if err != nil {
		if bz, err = base64.StdEncoding.DecodeString(in); err != nil {
			panic("pubkey must be 32-byte ed25519 in hex or base64")
		}
	}
	if len(bz) != 32 {
		panic(fmt.Sprintf("expected 32 bytes, got %d", len(bz)))
	}
	var pk ed25519.PubKeyEd25519
	copy(pk[:], bz)
	fmt.Println("gpub:   ", crypto.PubKeyToBech32(pk))
	fmt.Println("address:", pk.Address().String())
}
GO

go run ./pkconv.go 2C854661478AA1CDC954D11ABA6ABB6DBF469572564C24C61ABFC0622A04D350
# gpub:    gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9z...
# address: g1qmptf8uxdg6l0rh07jwvur0kk8my9vrdf5qtp4
```

### L4. Seed genesis with the Ledger validator

Because the validator key lives on the device, you **can't** use
`-lazy` (it would mint a throwaway local key and put *that* in genesis
instead). Build the genesis explicitly with `gnogenesis`
(`contribs/gnogenesis`), using the `gpub`/address from L3:

```sh
gnogenesis generate -chain-id "$CHAIN_ID" -output-path ./gnoland-data/genesis.json
gnogenesis validator add \
  -genesis-path ./gnoland-data/genesis.json \
  -address  g1qmptf8uxdg6l0rh07jwvur0kk8my9vrdf5qtp4 \
  -pub-key  gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9z... \
  -name     ledgervalidator \
  -power    10
```

`validator add` checks that the pubkey hashes to the address, so a
copy-paste slip is caught here.

You still need gnoland's node secrets for the **listener identity** —
run `secrets init` (step 1) as usual. Its `priv_validator_key.json` is
ignored for signing in tmkms mode; only `node_key.json` (the peer ID)
matters.

### L5. Configure gnoland and start (non-lazy)

Configure `tmkms_listener` exactly as in step 5, then start — with an
explicit `-genesis` and **no `-lazy`** (the genesis already exists):

```sh
# terminal 1: Ledger plugged in, app open
tmkms start -c ./tmkms/tmkms.toml

# terminal 2
"$GNOLAND" start \
  -data-dir ./gnoland-data \
  -genesis ./gnoland-data/genesis.json \
  -chainid "$CHAIN_ID" \
  -skip-genesis-sig-verification
```

Verify as in step 7 — gnoland should log `This node is a validator`
with the `gpub1…` from L3, the height should climb, and tmkms should log
`signed Proposal/Prevote/Precommit …`. The Tendermint Validator app
signs consensus messages without a per-vote button press (it must, for
liveness); it enforces its own monotonic HRS on-device, and tmkms's
`state_file` is still the authoritative gate — keep it.

> **Keep the device awake.** If the Ledger locks, sleeps, or the app is
> backgrounded, signing stops and the node stalls until you reopen the
> app. Plan for that on anything you care about.

## Where to go next

- **Run the repo's automated proof.** A build-tagged Go test drives a
  real tmkms binary through the full signing flow:

  ```sh
  go test -tags=tmkms_integration -count=1 -v ./tm2/pkg/bft/privval/upstream/...
  ```

- **Production hardening** — separate signer host, HSM/Ledger backend,
  TCP firewalling, `consensus.json` backups, Horcrux threshold signing:
  see [tmkms.md](tmkms.md), especially *Security notes*, *TCP vs UDS*,
  and the *Operational checklist*.
