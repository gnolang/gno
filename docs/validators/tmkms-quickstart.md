# gnoland + tmkms: secure setup

Stand up a gnoland validator whose consensus key lives in [tmkms], not on
the node. tmkms 0.15.0's risks are mostly **operational footguns** (bad
defaults, secret-leaking convenience commands, missing permission checks),
not accidental code bugs — this guide provisions around them. One residual
*code* risk remains (state-file durability on power loss); it's flagged
where it applies. See [tmkms.md](tmkms.md) for architecture and threat model.

[tmkms]: https://github.com/iqlusioninc/tmkms

## How to read this guide

Ordered real-stake-first:

- **[Prerequisites](#prerequisites--build-the-helper-tools)** — helper tools, binaries on PATH, GNOROOT.
- **[Part A](#part-a--host-hardening-do-this-first-for-every-backend)** — host hardening (users, dirs, transport); do first.
- **[Part B](#part-b--production-validator-with-a-yubihsm-recommended)** — YubiHSM, key never leaves hardware (recommended).
- **[Part C](#part-c--testnetlab-with-softsign-key-on-disk)** — softsign, key on disk (lab/learning only).
- **[Part D](#part-d--advanced-ledger-hardware-signer)** — Ledger hardware signer (advanced/low-stake).
- **[Part E](#part-e--prove-it-with-the-repos-automated-test)** + [checklist](#secure-setup-checklist).

> **Verified vs. derived.** The gnoland side of every path (listener,
> allowlist, peer-ID pin, `v0.34` pin, signing) is covered by the repo's
> real-tmkms test with softsign and a Ledger (Part E). Part B's YubiHSM
> steps are derived from the review + tmkms/YubiHSM docs — verify the exact
> tokens against your firmware; the policy they encode (non-exportable key,
> audit on, minimal caps) is firm.

## The security model in one paragraph

tmkms **dials** gnoland (gnoland listens); the connection is mutually
authenticated, and *how* depends on the transport
([A.3](#a3-choose-the-transport-unix-socket-or-firewalled-tcp)) — the
single most important choice here. Three ed25519 keys are in play — don't
mix them up:

| Key | Lives in | Role |
|---|---|---|
| **consensus key** | the signer (YubiHSM / Ledger / softsign file) | signs votes & proposals; its pubkey is the validator's identity in genesis |
| **kms-identity key** | tmkms `kms-identity.key` | tmkms's SecretConnection identity; its pubkey goes in gnoland's `allowed_kms_pubkeys` (TCP only) |
| **node key** | gnoland `node_key.json` | gnoland's SecretConnection identity; its peer ID (hex) is pinned in tmkms's `addr` (TCP only) |

`chain_id` must be **identical** everywhere (gnoland genesis,
`tmkms_listener.chain_id`, tmkms's `[[chain]].id` and
`[[validator]].chain_id`) or tmkms refuses to sign. Set once:

```sh
export CHAIN_ID=gno-tmkms-prod
```

---

# Prerequisites — build the helper tools

Two tiny **stdlib-only** Go helpers, compiled once. They build/run from
any directory (Part B.3's `pkconv.go` is the exception — it imports gno
crypto, run from the repo root; defined where it's used).

> **Run helpers as `gnoland`.** In Parts B/D the secrets live under
> `/var/lib` owned by `gnoland`; run each helper `sudo -u gnoland` (least
> privilege, no copying the secret to your dir). That user has no home, so
> the Go toolchain can't run as it: compile as yourself, run only the built
> binary under sudo.

`nodeid-hex.go` — prints a gnoland node's peer ID in the hex form tmkms
pins in its `addr` (TCP layouts):

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
```

`tmkms-identity-keygen.go` — generates tmkms's SecretConnection identity
key at the path you pass and prints its pubkey for gnoland's
`allowed_kms_pubkeys`:

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
	if err := os.WriteFile(os.Args[1], []byte(base64.StdEncoding.EncodeToString(seed)), 0o600); err != nil {
		panic(err)
	}
	fmt.Println("ed25519:" + hex.EncodeToString(pub)) // for gnoland's allowed_kms_pubkeys
}
GO
```

Compile as yourself, install where the service users can run them:

```sh
go build -o nodeid-hex ./nodeid-hex.go
go build -o tmkms-identity-keygen ./tmkms-identity-keygen.go
sudo install -m 0755 nodeid-hex tmkms-identity-keygen /usr/local/bin/
```

One-shot tools; `sudo rm /usr/local/bin/{nodeid-hex,tmkms-identity-keygen}`
afterward if you don't want them lying around.

## Put the binaries on a system path

Clone gno **once** to `/opt/gno`: you build the binaries from it and later
point `GNOROOT` at it, so the binary and its stdlibs always match. Own it as
your user (to build in place); it stays world-readable for the service
users, who — having no home — can't reach a clone under yours and otherwise
hit `Permission denied` on `sudo -u … <bin>`. Build gnoland and gnogenesis
and install them to `/usr/local/bin`:

```sh
sudo install -d -o "$(id -un)" /opt/gno
git clone https://github.com/gnolang/gno /opt/gno   # or reuse a clone, building from the same checkout
cd /opt/gno
make -C gno.land build.gnoland
make -C contribs/gnogenesis build
sudo install -m 0755 gno.land/build/gnoland contribs/gnogenesis/build/gnogenesis /usr/local/bin/
```

## Set `GNOROOT` to a gno checkout

gnoland loads stdlib `.gno` sources from `$GNOROOT/gnovm/stdlibs/` on every
start (and panics without `GNOROOT` set). Point it at the `/opt/gno`
checkout you just built from — same commit, so the stdlibs match the binary
(empty/mismatched fails with `failed loading stdlib "…": does not exist`).
It's world-readable, so the `gnoland` user can read it:

```sh
export GNOROOT=/opt/gno
```

`sudo -u gnoland` scrubs the env, so every `gnoland` command below passes
`env GNOROOT="$GNOROOT"`.

---

# Part A — Host hardening (do this first, for every backend)

tmkms doesn't enforce permissions on its own key/state/config (it accepts
world-readable secrets silently) — the OS must.

## A.1 Run everything as the gnoland user

Never run as root or your login user. Use **one** locked-down system
account, `gnoland`, for both gnoland and tmkms — on each host, if you run
them on two:

```sh
sudo useradd --system --no-create-home --shell /usr/sbin/nologin gnoland
```

## A.2 Lay out directories with strict permissions

The state file is the **double-sign gate** — losing or rolling it back can
get you slashed. Give the secrets a private `0600` home:

```sh
sudo install -d -o gnoland -g gnoland -m 700 /etc/tmkms          # tmkms config
sudo install -d -o gnoland -g gnoland -m 700 /var/lib/tmkms       # state + identity key
sudo install -d -o gnoland -g gnoland -m 700 /var/lib/tmkms/secrets
sudo install -d -o gnoland -g gnoland -m 700 /var/lib/gnoland     # gnoland data dir
```

- Secrets + state `0600`, owned by `gnoland` — tmkms won't reject
  world-readable secrets, so you must.
- **Absolute paths** in `tmkms.toml` — a CWD-relative state file can
  silently become a fresh, height-zero one.
- Password in a separate `password_file` (B.5), never inline — keeps it out
  of the config and process listings.

## A.3 Choose the transport: Unix socket or firewalled TCP

Two layouts, different auth boundaries — pick deliberately:

### Same host → Unix-domain socket (auth by filesystem)

gnoland creates the socket `0600` owned by `gnoland`; tmkms (also `gnoland`)
connects. Just create the dir:

```sh
sudo install -d -o gnoland -g gnoland -m 700 /run/gnoland
```

> On UDS the `allowed_kms_pubkeys` allowlist is **ignored** — gnoland does no
> SecretConnection on a unix socket, so there is no signer pubkey to check
> against. Leave it empty; the `0600` socket permission is the whole auth
> boundary. (If you set it anyway, gnoland logs that it is ignored.) The
> allowlist is only required, and only enforced, on a `tcp://` listener.

### TCP, firewalled (auth by cryptography)

Run the signer on its **own host** (or `127.0.0.1` on one box). Auth is
cryptographic, both halves mandatory:

- gnoland verifies tmkms via `allowed_kms_pubkeys` (the signer's pubkey).
- tmkms verifies gnoland via the **peer ID pinned in its `addr`**.

Firewall the port to the signer host only (crypto is the lock; the firewall
stops strangers knocking or using you as a signing oracle to probe):

```sh
sudo ufw allow from <SIGNER_HOST_IP> to any port 26659 proto tcp
sudo ufw deny 26659/tcp
```

Bind gnoland to the specific interface, not a wildcard, when you can.

Placeholders for the rest of the guide:

```sh
# UDS:
export PRIVVAL_LISTEN="unix:///run/gnoland/privval.sock"
export TMKMS_ADDR="unix:///run/gnoland/privval.sock"   # no peer-ID on UDS
# TCP (peer-ID prefix on TMKMS_ADDR is mandatory; filled at B.3):
export PRIVVAL_LISTEN="tcp://<validator_interface_ip>:26659"
```

Examples below show both: **Part B uses TCP**, **Parts C/D use UDS** — mix
transport and backend freely.

---

# Part B — Production validator with a YubiHSM (recommended)

A YubiHSM holds the consensus key in hardware and, provisioned correctly,
**never lets it out** — removing the entire "key stolen/exported off disk"
class softsign can't. The catch: tmkms's *convenience* tooling around the
YubiHSM is where the review found its sharpest edges. Rule for this
section: **provision by hand, never run the automated setup/test commands
against a production key.**

Install tmkms with the `yubihsm` provider (not in the default build):

```sh
cargo install tmkms --version 0.15.0 --features yubihsm --locked
sudo install -m 0755 ~/.cargo/bin/tmkms /usr/local/bin/   # so sudo -u gnoland can run it
tmkms version   # → 0.15.0
```

## B.1 Provision the YubiHSM by hand — never `tmkms yubihsm setup`

> **Do not run `tmkms yubihsm setup`.** That command prints the recovery
> mnemonic, the wrap key, and the auth passwords to standard output
> (captured by every terminal logger and scrollback), and it provisions
> roles with key-export capability. Provision with `yubihsm-shell`
> instead, so the secrets never leave your control.

Work on an **offline** machine. The policy to encode:

1. **Generate the consensus key inside the device** (or import once and
   destroy the source), **without** `exportable-under-wrap` — so a stolen
   credential can't wrap it out.
2. **tmkms's auth key gets `sign-eddsa` only** — no `export-wrapped`, no
   `sign-attestation-certificate`. If it leaks it can request signatures
   (still gated by firewall + state machine) but not clone the key.
3. **Force the audit log on** (ships off) — the HSM logs every op and
   refuses to sign once the buffer fills, so misuse leaves a trail.
4. **Store the recovery seed offline on paper** — never on the host.

`<tmkms-auth-password>` is a credential **you choose** for tmkms to log in.
Generate a long random one; keep it only in the `0600` `password_file`
below, never in `tmkms.toml`:

```sh
TMKMS_AUTH_PASSWORD=$(openssl rand -base64 32)   # generate once; use it below and in the password_file
```

A representative `yubihsm-shell` session (verify exact token spelling
against your firmware). Substitute `$TMKMS_AUTH_PASSWORD`:

```sh
yubihsm-shell
# connect
# session open 1 <default-or-changed-admin-password>

# Force the audit log on (records every op; refuses to sign when full):
# put option 0 force-audit 01

# Generate a NON-exportable ed25519 consensus key (note: no
# 'exportable-under-wrap' in the capability list):
# generate asymmetric 0 100 "gno-consensus" 1 sign-eddsa ed25519

# Create the auth key tmkms will use — sign-eddsa ONLY, no export:
# put authkey 0 2 "tmkms-signer" 1 sign-eddsa sign-eddsa <tmkms-auth-password>

# Replace/disable the default admin auth key (id 1) afterwards.
```

Never grant the auth key `export-wrapped` (turns a leak into key theft) or
`sign-attestation-certificate` (lets it forge attestations).

Store the password in a file, not the config:

```sh
printf '%s' "$TMKMS_AUTH_PASSWORD" | sudo tee /etc/tmkms/yubihsm-password >/dev/null
sudo chown gnoland:gnoland /etc/tmkms/yubihsm-password
sudo chmod 600 /etc/tmkms/yubihsm-password
unset TMKMS_AUTH_PASSWORD   # drop it from the shell once it's in the file
```

## B.2 Generate gnoland's node + listener identity

The key lives in the HSM, but gnoland still needs a node identity (the peer
ID tmkms pins) and tmkms needs a SecretConnection identity key. Generate
the node secrets:

```sh
sudo -u gnoland env GNOROOT="$GNOROOT" gnoland secrets init -data-dir /var/lib/gnoland/secrets
```

In tmkms mode `priv_validator_key.json` doesn't sign — only `node_key.json`
(the peer ID) matters.

**Peer ID in hex (TCP)** — tmkms pins it to sign only for *your* node.
gnoland reports bech32; tmkms wants the same 20 bytes in hex. Run
`nodeid-hex` as the gnoland user (it owns the key):

```sh
export VALIDATOR_PEER_ID=$(sudo -u gnoland nodeid-hex /var/lib/gnoland/secrets/node_key.json)
echo "$VALIDATOR_PEER_ID"   # e.g. 243cef06…dcac — pinned in B.4
```

**tmkms's SecretConnection identity (TCP)** — its pubkey must be in
`allowed_kms_pubkeys`. Run `tmkms-identity-keygen` as `gnoland` so the key
lands gnoland-owned:

```sh
ALLOW=$(sudo -u gnoland tmkms-identity-keygen /var/lib/tmkms/secrets/kms-identity.key)
echo "$ALLOW"   # e.g. ed25519:4b6efade…b18a — used in B.5
```

## B.3 Read the consensus pubkey out of the HSM and register the validator

The key never leaves the HSM, so you read its **public** half from tmkms
and register that. Start tmkms once (as `gnoland`, `tmkms.toml` from
B.4); it logs the consensus pubkey on startup. Read it, then `Ctrl-C`
(gnoland isn't up; tmkms just retries meanwhile):

```sh
sudo -u gnoland tmkms start -c /etc/tmkms/tmkms.toml
```

```
[keyring:yubihsm] added consensus Ed25519 key: 2C854661478AA1CDC954D11ABA6ABB6DBF469572564C24C61ABFC0622A04D350
```

That hex is an **example** — copy the one *your* device logs.

Convert it to gno's `gpub1…` / `g1…` forms with this helper, run **from the
gno repo root** (so the imports resolve):

```sh
cat > pkconv.go <<'GO'
// Converts a raw ed25519 consensus pubkey (hex or base64) to gno's
// gpub1… / g1… forms.
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

# pass the hex YOUR device logged (the value below is just an example):
go run ./pkconv.go 2C854661478AA1CDC954D11ABA6ABB6DBF469572564C24C61ABFC0622A04D350
# gpub:    gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9z...
# address: g1qmptf8uxdg6l0rh07jwvur0kk8my9vrdf5qtp4
```

That `gpub1…` / `g1…` pair is your validator's public identity. Then:

**Joining an existing chain (production).** Don't generate genesis — submit
the `gpub1…` / `g1…` to the chain's validator-onboarding path. Keep the
pair (confirmed in B.7), then skip to [B.4](#b4-write-tmkmstoml).

**Bootstrapping your own chain (test).** Build genesis with `gnogenesis`,
seeded with the HSM pubkey. **No** `gnoland start -lazy` — it would mint a
throwaway local key and seed genesis with that instead:

```sh
sudo -u gnoland gnogenesis generate -chain-id "$CHAIN_ID" -output-path /var/lib/gnoland/genesis.json
sudo -u gnoland gnogenesis validator add \
  -genesis-path /var/lib/gnoland/genesis.json \
  -address  g1qmptf8uxdg6l0rh07jwvur0kk8my9vrdf5qtp4 \
  -pub-key  gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9z... \
  -name     hsm-validator \
  -power    10
```

`validator add` checks the pubkey hashes to the address, catching a
copy-paste slip here (same check the onboarding path runs).

## B.4 Write `tmkms.toml`

Set the signer `addr` to match A.3 — pick **one** (the TCP form embeds the
mandatory peer-ID pin from B.2):

```sh
# TCP (separate hosts):
export TMKMS_ADDR="tcp://${VALIDATOR_PEER_ID}@<validator_ip>:26659"
# OR same host (Unix socket — no peer ID; socket perms are the boundary):
export TMKMS_ADDR="unix:///run/gnoland/privval.sock"
```

Write the config, then lock it `0600` owned by `gnoland` (password stays in
the `password_file`; all paths absolute):

```sh
sudo tee /etc/tmkms/tmkms.toml >/dev/null <<TOML
[[chain]]
id = "${CHAIN_ID}"
key_format = { type = "hex" }                                # logs the pubkey in hex (B.3)
state_file = "/var/lib/tmkms/secrets/consensus_state.json"   # absolute! the HRS gate

[[providers.yubihsm]]
adapter = { type = "usb" }                                   # or { type = "http", addr = "..." }
auth = { key = 2, password_file = "/etc/tmkms/yubihsm-password" }
keys = [{ chain_ids = ["${CHAIN_ID}"], key = 100, type = "consensus" }]
# serial_number = "0123456789"                               # pin to one physical device

[[validator]]
chain_id = "${CHAIN_ID}"
addr = "${TMKMS_ADDR}"                                       # from above; peer-ID pin is MANDATORY on TCP
secret_key = "/var/lib/tmkms/secrets/kms-identity.key"
protocol_version = "v0.34"
reconnect = true
TOML

sudo chown gnoland:gnoland /etc/tmkms/tmkms.toml
sudo chmod 600 /etc/tmkms/tmkms.toml
```

Security-weight fields:

- **`addr` peer-ID pin (TCP)** — the `<peer_id>@` prefix lets tmkms verify
  gnoland. Omit it and tmkms signs for whoever answers the port (only
  logging `unverified validator peer ID!`). Always pin `$VALIDATOR_PEER_ID`.
  (No peer ID on `unix://`; socket perms are the boundary.)
- **`state_file` (absolute, durable)** — the double-sign (HRS) gate. Never
  restore a stale copy (classic self-double-sign). One caveat in B.7.
- **`password_file`, not inline `password`** — keeps the credential out of
  the config and process listings.
- **`protocol_version = "v0.34"`** — gnoland speaks only v0.34; the
  `deprecated … update to v0.38!` warning is expected. Don't drop or bump it.
- **No `state_hook`** — its stdout is never captured, so a `fail_closed`
  hook can silently kill startup. Seed the state file directly.

## B.5 Configure gnoland's tmkms listener

Set the four `tmkms_listener` fields. The allowlist must be non-empty — an
empty one is fail-open (accept any peer that handshakes). Run as the
gnoland user (all `gnoland config` commands write under `/var/lib/gnoland`):

```sh
CFG=/var/lib/gnoland/config/config.toml
sudo -u gnoland env GNOROOT="$GNOROOT" gnoland config init -config-path "$CFG"
```

> **Ordering gotcha.** A non-empty `listen_addr` enables validation of the
> whole block, so set it **last** — set it first and the write is rejected
> (`chain_id` still empty) and silently stays unset.

```sh
sudo -u gnoland env GNOROOT="$GNOROOT" gnoland config set -config-path "$CFG" consensus.priv_validator.tmkms_listener.chain_id "$CHAIN_ID"
sudo -u gnoland env GNOROOT="$GNOROOT" gnoland config set -config-path "$CFG" consensus.priv_validator.tmkms_listener.protocol_version "v0.34"
sudo -u gnoland env GNOROOT="$GNOROOT" gnoland config set -config-path "$CFG" consensus.priv_validator.tmkms_listener.allowed_kms_pubkeys "$ALLOW"
# listen_addr LAST — this enables the mode:
sudo -u gnoland env GNOROOT="$GNOROOT" gnoland config set -config-path "$CFG" consensus.priv_validator.tmkms_listener.listen_addr "$PRIVVAL_LISTEN"
```

Verify it stuck (`listen_addr` should be your `$PRIVVAL_LISTEN`, not
empty):

```sh
sudo -u gnoland env GNOROOT="$GNOROOT" gnoland config get -config-path "$CFG" consensus.priv_validator.tmkms_listener
```

## B.6 Run both, in the right order

**Order matters.** gnoland blocks up to `wait_for_connection_timeout`
(default 60s) waiting for tmkms, so start tmkms first (`reconnect = true`
makes it retry). Run each in its own terminal, both as `gnoland`:

```sh
# signer (as gnoland)
sudo -u gnoland tmkms start -c /etc/tmkms/tmkms.toml

# validator (as the gnoland user) — -genesis = the network's genesis;
# NO -lazy (you're not minting a local key).
sudo -u gnoland env GNOROOT="$GNOROOT" gnoland start \
  -data-dir /var/lib/gnoland \
  -genesis /var/lib/gnoland/genesis.json \
  -chainid "$CHAIN_ID"
```

## B.7 Verify it's really the HSM signing — and keep it that way

Within seconds the node should advance through heights.

- **tmkms log:** `connected to validator successfully`, then ongoing
  `signed Proposal/Prevote/Precommit …`. **No** `unverified validator
  peer ID!` line — if you see that, the peer-ID pin in `addr` (B.4) is
  missing or wrong; tmkms prints what it expected, compare to
  `$VALIDATOR_PEER_ID`.
- **gnoland log:** `This node is a validator` with the `gpub1…` from B.3,
  and `Committed state … height=N` climbing.
- **Identity check:** the validator address in the `Signed and pushed
  vote` lines must equal the address from B.3 and the genesis entry.
- **Kill tmkms** and watch gnoland stop committing; restart it and watch
  it resume. That is the proof the key lives in the HSM, not the node:

  ```sh
  sudo pkill -f 'tmkms start'   # gnoland's committed height stops climbing...
  sudo -u gnoland tmkms start -c /etc/tmkms/tmkms.toml   # ...resumes after tmkms reconnects
  ```

> **Benign log noise.** `SignerListener: accept failed … i/o timeout` and
> `already connected, dropping listen request` are idle re-accept attempts
> on the one live connection — ignore them, watch the committed height.

**Keep it safe in production:**

- **Residual code risk — state-file durability.** tmkms 0.15.0 doesn't
  `fsync` its state file before signing the next block, so a power loss can
  roll the on-disk height backward → double-sign. Mitigate: reliable power
  (UPS), durable local storage (no flaky network mount), and **never**
  restore an old state file after a crash. (A two-line upstream `fsync`
  patch closes it.)
- **Back up the state file** (for recovery of the *latest* state, not as a
  rollback checkpoint):

  ```sh
  sudo install -o gnoland -g gnoland -m 600 \
    /var/lib/tmkms/secrets/consensus_state.json \
    /var/lib/tmkms/secrets/consensus_state.json.bak
  ```
- **Drain the audit log** — with forced auditing on (B.1) the HSM stops
  signing when the buffer fills; run an auditor that reads/clears it (~30s)
  and ships it off-box.
- **Never `tmkms yubihsm test` against the production key** — it's an
  unbounded signing oracle that bypasses the double-sign gate.
- **Don't re-run provisioning over a live key** — the key-write path has no
  "refuse if exists" guard and can silently overwrite. Provision once.

---

# Part C — Testnet/lab with softsign (key on disk)

softsign keeps the consensus key in a **file on the host** — fine for
testnets and learning the wiring, but **not for real stake** (use Part B).
The *connection* (transport, allowlist, peer-ID pin, `protocol_version`,
verification) is identical to Part B; only key custody and genesis differ.

This lab is a same-host **UDS** layout, so per A.1 it runs **both processes
as your operator user** in local dirs (no system users). softsign-specific
hardening:

- **`0600` on `consensus.key` and the state file** — a readable softsign
  key *is* the validator's private key.
- **Disable swap** — tmkms zeroizes the key but doesn't `mlock` it, so it
  can be paged to disk:

  ```sh
  sudo swapoff -a                     # and comment swap out of /etc/fstab to persist
  ```
- **Don't re-run `secrets init` / keygen over an existing key** — it can
  overwrite without warning.

## C.1 Generate node secrets and export the consensus key

Install tmkms with the softsign feature (not in the default build):

```sh
cargo install tmkms --version 0.15.0 --features softsign --locked
```

Generate the node secrets (this also makes the consensus key softsign
loads and genesis registers):

```sh
gnoland secrets init -data-dir ./gnoland-data/secrets
```

gnoland stores the key as base64 of the 64-byte `seed‖pubkey`; softsign
wants base64 of just the 32-byte **seed**. Reslice (just bytes, plain
`python3`):

```sh
mkdir -p ./tmkms/secrets
python3 - <<'PY' > ./tmkms/secrets/consensus.key
import json, base64
v = json.load(open("gnoland-data/secrets/priv_validator_key.json"))["priv_key"]["value"]
print(base64.b64encode(base64.b64decode(v)[:32]).decode(), end="")
PY
chmod 600 ./tmkms/secrets/consensus.key
```

UDS layout, so **no peer-ID pin** and **no allowlist** — gnoland ignores
`allowed_kms_pubkeys` on a unix socket (A.3). tmkms still wants a
`secret_key` file for its `[[validator]]` block, so generate one; its pubkey
is not registered anywhere on UDS:

```sh
tmkms-identity-keygen ./tmkms/secrets/kms-identity.key >/dev/null   # operator-owned, no sudo
```

## C.2 tmkms.toml with the softsign provider

Like B.4, but **softsign** provider and a **Unix socket** (no peer-ID pin).
tmkms needs **absolute** paths (A.2), so resolve the lab dirs first:

```sh
export TMKMS_DIR=$(cd ./tmkms && pwd)                            # ./tmkms from C.1
export PRIVVAL_SOCK="$(cd ./gnoland-data && pwd)/privval.sock"   # gnoland creates it at 0600

tee "$TMKMS_DIR/tmkms.toml" >/dev/null <<TOML
[[chain]]
id = "gno-tmkms-test"
key_format = { type = "hex" }
state_file = "$TMKMS_DIR/secrets/consensus_state.json"

[[providers.softsign]]
chain_ids = ["gno-tmkms-test"]
key_type = "consensus"
key_format = "base64"
path = "$TMKMS_DIR/secrets/consensus.key"

[[validator]]
chain_id = "gno-tmkms-test"
addr = "unix://$PRIVVAL_SOCK"   # same-host UDS; no peer-ID pin (A.3)
secret_key = "$TMKMS_DIR/secrets/kms-identity.key"
protocol_version = "v0.34"
reconnect = true
TOML
chmod 600 "$TMKMS_DIR/tmkms.toml"
```

## C.3 Configure gnoland's tmkms listener and start (lazy genesis is fine here)

The key is on disk *and* registered locally, so `-lazy` can build genesis
from `priv_validator_key.json`.

Configure the listener. On UDS you set three `tmkms_listener` fields —
`allowed_kms_pubkeys` stays empty because gnoland ignores it on a socket
(the `0600` socket perms are the real boundary, A.3):

```sh
export CHAIN_ID=gno-tmkms-test                 # must match C.2's [[chain]].id
export PRIVVAL_LISTEN="unix://$PRIVVAL_SOCK"   # gnoland listens on the socket from C.2
CFG=./gnoland-data/config/config.toml
gnoland config init -config-path "$CFG"
```

> **Ordering gotcha.** A non-empty `listen_addr` enables validation of the
> whole block, so set it **last** — set it first and the write is rejected
> (`chain_id` still empty) and silently stays unset.

```sh
gnoland config set -config-path "$CFG" consensus.priv_validator.tmkms_listener.chain_id "$CHAIN_ID"
gnoland config set -config-path "$CFG" consensus.priv_validator.tmkms_listener.protocol_version "v0.34"
# listen_addr LAST — this enables the mode:
gnoland config set -config-path "$CFG" consensus.priv_validator.tmkms_listener.listen_addr "$PRIVVAL_LISTEN"
```

Verify it stuck (`listen_addr` should be your `$PRIVVAL_LISTEN`, not
empty):

```sh
gnoland config get -config-path "$CFG" consensus.priv_validator.tmkms_listener
```

Both run as your operator user, so the exported `GNOROOT` carries through
(no `env` prefix). `-lazy` also reads `examples/` from `GNOROOT` — another
reason it must be a full checkout. Start both:

```sh
# terminal 1
tmkms start -c ./tmkms/tmkms.toml

# terminal 2
gnoland start \
  -data-dir ./gnoland-data \
  -genesis ./gnoland-data/genesis.json \
  -chainid "$CHAIN_ID" \
  -lazy \
  -skip-genesis-sig-verification
```

- `-skip-genesis-sig-verification` — standard dev-genesis flag (the `-lazy`
  genesis has example txs from test accounts); unrelated to tmkms.
- `-lazy` only generates *missing* files (won't clobber tmkms config), reads
  the local pubkey to register the validator, then signs via tmkms.

Verify as in B.7.

---

# Part D — Advanced: Ledger hardware signer

A Ledger holds the consensus key **on the device**, like a YubiHSM but a
single signer with no failover — so advanced/low-stake, not HA mainnet (use
a YubiHSM or Horcrux for HA). Same "read the pubkey out, seed genesis" flow
as Part B; only the provider and device handling change.

> **Residual hardware caveat.** tmkms 0.15.0 doesn't check the Ledger's
> APDU return codes, so a device/transport error could be accepted as
> valid. In practice the `gnogenesis validator add` pubkey→address check
> (B.3) catches a wrong pubkey, and a bad signature just fails consensus.
> Use a known-good device and verify the pubkey out-of-band.

## D.1 Prerequisites

Same-host **UDS** layout; tmkms runs as `gnoland` like everything else
(A.1).

Install tmkms with the Ledger provider — the feature is `ledger` in 0.15.0
(the config block is still `[[providers.ledgertm]]`); 0.15.0 has no default
features:

```sh
cargo install tmkms --version 0.15.0 --features ledger --locked
# or, to keep softsign too: --features ledger,softsign
sudo install -m 0755 ~/.cargo/bin/tmkms /usr/local/bin/   # so sudo -u gnoland can run it
```

You also need:

- A Ledger with the **"Tendermint Validator"** app (the ed25519 consensus
  app, *not* the Cosmos app) plugged in, unlocked, app **open**.
- **Device access for `gnoland`.** If `gnoland` can't open the Ledger's
  `/dev/hidraw*`, `tmkms start` fails with `signing operation failed`. Grant
  the `gnoland` group the hidraw device, reload, then **replug**:

  ```sh
  echo 'KERNEL=="hidraw*", SUBSYSTEM=="hidraw", ATTRS{idVendor}=="2c97", MODE="0660", GROUP="gnoland"' \
    | sudo tee /etc/udev/rules.d/51-gnoland-ledger.rules
  sudo udevadm control --reload-rules && sudo udevadm trigger
  ```

## D.2 Generate gnoland's node and listener identity

The key is on the Ledger, but gnoland still needs a node identity and tmkms
a SecretConnection identity key. Generate the node secrets:

```sh
sudo -u gnoland env GNOROOT="$GNOROOT" gnoland secrets init -data-dir /var/lib/gnoland/secrets
```

`priv_validator_key.json` doesn't sign (the Ledger does), and UDS has **no
peer-ID pin** and **no allowlist** (gnoland ignores `allowed_kms_pubkeys` on
a socket, A.3). tmkms still wants a `secret_key` file for its
`[[validator]]` block, so generate one as `gnoland` (its pubkey is not
registered anywhere on UDS):

```sh
sudo -u gnoland tmkms-identity-keygen /var/lib/tmkms/secrets/kms-identity.key >/dev/null
```

## D.3 Write `tmkms.toml`

Socket dir from [A.3](#a3-choose-the-transport-unix-socket-or-firewalled-tcp)
(create it if you skipped it):

```sh
sudo install -d -o gnoland -g gnoland -m 700 /run/gnoland
```

Write the config, lock it `0600` owned by `gnoland`. Provider is `ledgertm`
(no key files — the key is on the device); UDS, so no peer-ID pin:

```sh
sudo tee /etc/tmkms/tmkms.toml >/dev/null <<TOML
[[chain]]
id = "${CHAIN_ID}"
key_format = { type = "hex" }                                # logs the pubkey in hex (D.4)
state_file = "/var/lib/tmkms/secrets/consensus_state.json"   # absolute! the HRS gate

[[providers.ledgertm]]
chain_ids = ["${CHAIN_ID}"]

[[validator]]
chain_id = "${CHAIN_ID}"
addr = "unix:///run/gnoland/privval.sock"                    # same-host UDS; no peer-ID pin (A.3)
secret_key = "/var/lib/tmkms/secrets/kms-identity.key"
protocol_version = "v0.34"
reconnect = true
TOML

sudo chown gnoland:gnoland /etc/tmkms/tmkms.toml
sudo chmod 600 /etc/tmkms/tmkms.toml
```

One `[[providers.ledgertm]]`, always signs the **consensus** key.
Security-weight fields:

- **`state_file` (absolute)** — the double-sign (HRS) gate. Never restore a
  stale copy (D.6).
- **`protocol_version = "v0.34"`** — gnoland speaks only v0.34; the
  `deprecated … update to v0.38!` warning is expected. Don't drop or bump it.
- **No peer-ID pin / no `state_hook`** — UDS has no `<peer_id>@`; and a
  `state_hook`'s stdout is never captured, so a `fail_closed` hook can
  silently kill startup.

## D.4 Read the consensus pubkey off the device and register the validator

The key never leaves the Ledger, so read its **public** half from tmkms.
Start tmkms once (as `gnoland`, `tmkms.toml` from D.3); it logs the consensus
pubkey on startup. Read it, then `Ctrl-C` (gnoland isn't up; tmkms retries):

```sh
sudo -u gnoland tmkms start -c /etc/tmkms/tmkms.toml
```

```
[keyring:ledgertm] added consensus Ed25519 key: 2C854661478AA1CDC954D11ABA6ABB6DBF469572564C24C61ABFC0622A04D350
```

That hex is an **example** — copy the one *your* device logs.

Convert it to gno's `gpub1…` / `g1…` forms with `pkconv.go`, run **from the
gno repo root**:

```sh
cat > pkconv.go <<'GO'
// Converts a raw ed25519 consensus pubkey (hex or base64) to gno's
// gpub1… / g1… forms.
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

# pass the hex YOUR device logged (the value below is just an example):
go run ./pkconv.go 2C854661478AA1CDC954D11ABA6ABB6DBF469572564C24C61ABFC0622A04D350
# gpub:    gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9z...
# address: g1qmptf8uxdg6l0rh07jwvur0kk8my9vrdf5qtp4
```

That `gpub1…` / `g1…` pair is your validator's identity. Either submit it to
an existing chain's onboarding path (no genesis to generate), or bootstrap
your own by seeding genesis with it. **No** `gnoland start -lazy` — it would
mint a throwaway key and seed genesis with that:

```sh
sudo -u gnoland gnogenesis generate -chain-id "$CHAIN_ID" -output-path /var/lib/gnoland/genesis.json
sudo -u gnoland gnogenesis validator add \
  -genesis-path /var/lib/gnoland/genesis.json \
  -address  g1qmptf8uxdg6l0rh07jwvur0kk8my9vrdf5qtp4 \
  -pub-key  gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9z... \
  -name     ledger-validator \
  -power    10
```

`validator add` checks that the pubkey hashes to the address, catching a
copy-paste slip right here.

## D.5 Configure gnoland's tmkms listener

Set three `tmkms_listener` fields. On UDS `allowed_kms_pubkeys` stays empty
because gnoland ignores it on a socket (A.3). Run as the gnoland user:

```sh
export PRIVVAL_LISTEN="unix:///run/gnoland/privval.sock"   # gnoland listens on the socket from D.3
CFG=/var/lib/gnoland/config/config.toml
sudo -u gnoland env GNOROOT="$GNOROOT" gnoland config init -config-path "$CFG"
```

> **Ordering gotcha.** A non-empty `listen_addr` enables validation of the
> whole block, so set it **last** — set it first and the write is rejected
> (`chain_id` still empty) and silently stays unset.

```sh
sudo -u gnoland env GNOROOT="$GNOROOT" gnoland config set -config-path "$CFG" consensus.priv_validator.tmkms_listener.chain_id "$CHAIN_ID"
sudo -u gnoland env GNOROOT="$GNOROOT" gnoland config set -config-path "$CFG" consensus.priv_validator.tmkms_listener.protocol_version "v0.34"
# listen_addr LAST — this enables the mode:
sudo -u gnoland env GNOROOT="$GNOROOT" gnoland config set -config-path "$CFG" consensus.priv_validator.tmkms_listener.listen_addr "$PRIVVAL_LISTEN"
```

Verify it stuck (`listen_addr` should be your socket, not empty):

```sh
sudo -u gnoland env GNOROOT="$GNOROOT" gnoland config get -config-path "$CFG" consensus.priv_validator.tmkms_listener
```

## D.6 Run both, verify, and keep the device awake

**Order matters.** gnoland blocks up to `wait_for_connection_timeout`
(default 60s) waiting for tmkms, so start tmkms first (`reconnect = true`
retries). Run each in its own terminal, both as `gnoland`:

```sh
# signer (as gnoland) — Ledger plugged in, unlocked, app open
sudo -u gnoland tmkms start -c /etc/tmkms/tmkms.toml

# validator (as gnoland) — NO -lazy; genesis already holds the
# device's pubkey from D.4
sudo -u gnoland env GNOROOT="$GNOROOT" gnoland start \
  -data-dir /var/lib/gnoland \
  -genesis /var/lib/gnoland/genesis.json \
  -chainid "$CHAIN_ID"
```

Within a few seconds the node should advance through heights:

- **tmkms log:** `connected to validator successfully`, then ongoing
  `signed Proposal/Prevote/Precommit …`.
- **gnoland log:** `This node is a validator` with the `gpub1…` from D.4,
  and `Committed state … height=N` climbing.
- **Identity check:** the validator address in the `Signed and pushed
  vote` lines must equal the address from D.4 and the genesis entry.
- **Kill tmkms** and watch gnoland stop committing; restart it and watch
  it resume — proof the key lives on the device, not the node.

> **Keep the device awake.** If the Ledger locks/sleeps or the app is
> backgrounded, signing stops and the node stalls until you reopen it.

> **State-file durability.** As in every backend, tmkms 0.15.0 doesn't
> `fsync` its state file before signing, so a power loss can roll the height
> backward (double-sign). Reliable power, never restore an old state file
> ([B.7](#b7-verify-its-really-the-hsm-signing--and-keep-it-that-way)).

---

# Part E — Prove it with the repo's automated test

A build-tagged Go test drives a **real tmkms binary** (softsign + Ledger)
through the full signing flow, exercising the gnoland side of every path:

```sh
go test -tags=tmkms_integration -count=1 -v ./tm2/pkg/bft/privval/upstream/...
```

| Symptom | Cause | Fix |
|---|---|---|
| `sudo -u gnoland …`: `unable to execute …: Permission denied` | binary under `~/.cargo/bin` or your home; the `gnoland` user can't traverse it | `sudo install -m 0755 <binary> /usr/local/bin/` ([Prerequisites](#put-the-binaries-on-a-system-path)) |
| gnoland panics: `unable to determine GNOROOT` | `GNOROOT` unset, or scrubbed by `sudo -u` | `export GNOROOT=/opt/gno` and pass `env GNOROOT="$GNOROOT"` through sudo ([Prerequisites](#set-gnoroot-to-a-gno-checkout)) |
| gnoland panics at InitChainer: `failed loading stdlib "…": does not exist` | `GNOROOT` isn't a complete checkout, or its stdlibs don't match the binary | Point `GNOROOT` at a full checkout of the tree you built, at the same commit ([Prerequisites](#set-gnoroot-to-a-gno-checkout)) |
| tmkms: `unknown field 'softsign', expected 'ledgertm'` | tmkms built without softsign | `cargo install tmkms --version 0.15.0 --features softsign --locked` |
| gnoland: `allowed_kms_pubkeys must not be empty` | allowlist unset, or `listen_addr` set before the other fields | Set the four fields with `listen_addr` **last** (B.5) |
| gnoland exits after ~60s waiting for connection | tmkms not running, can't reach `addr`, or pubkey not in allowlist | Start tmkms first; check `addr` vs `listen_addr`; confirm the hex in `allowed_kms_pubkeys` |
| `protocol_version` rejected at startup | a value other than `v0.34` | Use `v0.34` on both sides |
| gnoland panics at boot: `PubKey does not match Signer address` | default `-lazy` genesis has example txs (softsign path) | Add `-skip-genesis-sig-verification` (C.3) |
| Node never advances past height 1, no signing in tmkms log | `chain_id` mismatch across genesis / config.toml / tmkms.toml | Make all three identical and restart |
| Signatures produced but rejected by the chain | consensus key in the signer ≠ genesis validator key | Re-seed genesis from the *same* key the signer holds |
| tmkms: `unverified validator peer ID! (<hex>)` | peer-ID prefix missing from `addr` (TCP) | Pin `tcp://<hex>@host:port` with `$VALIDATOR_PEER_ID` (B.4) |
| tmkms (UDS): `I/O error: Permission denied (os error 13)` | tmkms isn't running as `gnoland`; the socket is `0600` owned by `gnoland` | Run tmkms as `gnoland` (A.1) |
| tmkms (Ledger): `error loading configuration: signing operation failed` | the user running tmkms can't open the Ledger's `/dev/hidraw*` node | Add the hidraw udev rule for that user's group, reload, replug (D.1) |
| tmkms (Ledger): startup `SigningError` / no pubkey | device locked, asleep, or wrong app | Unlock, open **Tendermint Validator** app, restart tmkms |

---

# Secure-setup checklist

Run down this list before putting real stake behind the validator. Each
item plainly states why it matters.

**Host & OS**
- [ ] gnoland and tmkms run as the dedicated, no-shell `gnoland` user —
  *limits what a compromised process can reach.*
- [ ] Config, state, and key files are `0600`, owned by `gnoland` — *tmkms
  won't reject world-readable secrets; the OS must.*
- [ ] All paths in `tmkms.toml` are absolute — *stops a CWD-relative
  state file from silently becoming a fresh, height-zero one.*

**Transport & auth**
- [ ] Same host: socket `0600` owned by `gnoland` — *on UDS, the owner-only
  socket is the auth boundary.*
- [ ] Separate hosts: TCP with the peer-ID pinned in `addr` **and** the
  signer pubkey in `allowed_kms_pubkeys` — *both halves of mutual auth;
  without the pin tmkms signs for any impostor on the port.*
- [ ] Listen port firewalled to the signer host only — *keeps strangers
  from knocking or probing your signer.*
- [ ] On TCP, `allowed_kms_pubkeys` is non-empty — *an empty list is
  fail-open. (On UDS it's ignored; leave it empty and rely on socket perms.)*

**Key custody (production)**
- [ ] Consensus key generated **non-exportable** in a YubiHSM; never run
  `tmkms yubihsm setup` (leaks secrets to stdout) — *the key can never be
  stolen off disk or wrapped out.*
- [ ] HSM auth key has `sign-eddsa` only — no `export-wrapped`, no
  attestation — *a leaked credential can request signatures but not clone
  the key.*
- [ ] Forced audit log on, drained by an auditor — *every op is recorded;
  misuse can't run unbounded.*
- [ ] Recovery seed stored offline on paper — *off the validator host.*

**State & operations**
- [ ] Signer on reliable power; state file on durable storage; never
  restore a stale state file — *mitigates the one residual code risk
  (no fsync before sign) that could cause a double-sign.*
- [ ] State file backed up as part of maintenance — *for recovery, not
  rollback.*
- [ ] Never run `tmkms yubihsm test` against the production key — *it's an
  unbounded oracle that bypasses the double-sign gate.*
- [ ] `protocol_version = "v0.34"` on both sides; stay on a tmkms release
  that supports it — *gnoland speaks only v0.34.*
- [ ] No `state_hook` configured — *the hook subsystem can silently kill
  startup; seed the state file directly.*
