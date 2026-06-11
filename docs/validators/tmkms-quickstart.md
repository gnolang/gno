# gnoland + tmkms: secure setup

A step-by-step guide to standing up a gnoland validator whose consensus
key is held by [tmkms] instead of living on the node — and configuring
both the host and tmkms so that the failure modes a security review of
tmkms surfaced are closed *before* they can bite you.

The review of tmkms 0.15.0 catalogued dozens of findings; the large
majority of them are not code bugs you can hit by accident but
**operational footguns** — bad defaults, convenience commands that leak
secrets, missing permission checks — every one of which is neutralised by
provisioning and configuring the system correctly. This guide bakes that
hardening in. Where a residual *code-level* risk remains (there is
essentially one that matters: state-file durability on power loss), it is
called out at the exact step where it applies.

[tmkms]: https://github.com/iqlusioninc/tmkms

## How to read this guide

It is ordered by what you should actually run for real stake first, and
descends toward test/lab setups:

- **[Prerequisites — helper tools](#prerequisites--build-the-helper-tools)** —
  two small stdlib-only Go helpers every backend uses; build them once.
- **[Part A — Host hardening](#part-a--host-hardening-do-this-first-for-every-backend)** —
  the OS-level groundwork (dedicated user, locked-down directories,
  transport choice). Do this regardless of which signer backend you pick.
- **[Part B — Production with a YubiHSM](#part-b--production-validator-with-a-yubihsm-recommended)** —
  the recommended production path. The consensus key is generated inside
  the HSM and **never exists outside it**.
- **[Part C — Testnet/lab with softsign](#part-c--testnetlab-with-softsign-key-on-disk)** —
  the file-based backend. Simple, but the key touches disk: testnets and
  learning only.
- **[Part D — Ledger hardware signer](#part-d--advanced-ledger-hardware-signer)** —
  a hardware key for advanced/low-stake setups without HSM-grade failover.
- **[Part E — Prove it](#part-e--prove-it-with-the-repos-automated-test)** and the
  **[checklist](#secure-setup-checklist)**.

> **What is verified vs. derived.** The gnoland side of every path here —
> the listener, the allowlist, the peer-ID pin, the `v0.34` protocol pin,
> and signing through tmkms — is exercised end-to-end by the repo's
> real-tmkms integration test, with **softsign** and a **Ledger** as the
> backends (see [Part E](#part-e--prove-it-with-the-repos-automated-test)).
> The **YubiHSM** provisioning and provider config in Part B are derived
> from the tmkms security review, the tmkms source, and YubiHSM's own
> documentation; they need the physical device to exercise. Treat the
> exact `yubihsm-shell` / TOML tokens as a recipe to verify against your
> firmware and tmkms version — but the *security policy* they encode
> (non-exportable key, audit log on, minimal capabilities) is firm.

See [tmkms.md](tmkms.md) for the architecture, the threat model, and the
operational checklist this guide is the hands-on companion to.

## The security model in one paragraph

The connection between gnoland and tmkms is **mutually authenticated**,
and the defining property of this mode is that **tmkms dials gnoland** —
gnoland listens, tmkms connects in. gnoland verifies it is talking to
*your* signer (not an impostor) and tmkms verifies it is talking to
*your* validator (not an attacker's listener). How that mutual check is
enforced depends on the transport, and getting it right is the single
most important configuration decision — see
[Part A.3](#a3-choose-the-transport-unix-socket-or-firewalled-tcp).

Three distinct ed25519 keys are in play — don't mix them up:

| Key | Lives in | Role |
|---|---|---|
| **consensus key** | the signer (YubiHSM / Ledger / softsign file) | signs votes & proposals; its pubkey is the validator's identity in genesis |
| **kms-identity key** | tmkms `kms-identity.key` | tmkms's SecretConnection identity; its pubkey goes in gnoland's `allowed_kms_pubkeys` (TCP only) |
| **node key** | gnoland `node_key.json` | gnoland's SecretConnection identity; its peer ID (hex) is pinned in tmkms's `addr` (TCP only) |

The `chain_id` must be **identical** everywhere — the gnoland genesis,
gnoland's `tmkms_listener.chain_id`, and tmkms's `[[chain]].id` /
`[[validator]].chain_id`. If they drift, tmkms refuses to sign. Pick it
once and reuse it:

```sh
export CHAIN_ID=gno-tmkms-prod
```

---

# Prerequisites — build the helper tools

Every backend below needs two tiny **stdlib-only** Go helpers. Write them
once, here, and later sections just `go run` them. They have no module
dependencies, so they run from any directory (Part B.3's `pkconv.go` is
the one exception — it imports the gno crypto packages and must be run
from the gno repo root; it's defined where it's used).

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

---

# Part A — Host hardening (do this first, for every backend)

These steps apply whether you end up on a YubiHSM, a Ledger, or softsign.
tmkms does **not** enforce file permissions on its own key, state, or
config files (it accepts world-readable secrets silently), so the OS must
enforce them for it.

## A.1 Run tmkms as its own locked-down user

Never run the signer as root or as your login user. A dedicated system
account with no shell and no home limits what a compromised tmkms process
— or anyone who later reads its files — can reach.

```sh
sudo useradd --system --no-create-home --shell /usr/sbin/nologin tmkms
```

If gnoland and tmkms share a host, run gnoland as its own user too (e.g.
`gnoland`). Two unprivileged users that can only meet over a single
tightly-permissioned socket is the local isolation you want.

## A.2 Lay out directories with strict permissions

tmkms's secrets and its state file are the crown jewels. The state file
in particular is the **double-sign gate** — losing or rolling it back is
the one thing that can get you slashed. Give them a private home:

```sh
sudo install -d -o tmkms -g tmkms -m 700 /etc/tmkms          # config
sudo install -d -o tmkms -g tmkms -m 700 /var/lib/tmkms       # state + identity key
sudo install -d -o tmkms -g tmkms -m 700 /var/lib/tmkms/secrets
```

Rules that close whole classes of findings at once:

- **Every secret and the state file is `0600`, owned by `tmkms`.** A
  world- or group-readable consensus key, identity key, or YubiHSM
  password is game over; tmkms won't warn you, so enforce it yourself.
- **Use absolute paths everywhere** in `tmkms.toml`. tmkms resolves some
  paths (notably the default state-file path) relative to its *current
  working directory*; an absolute path removes any ambiguity about which
  state file is the authoritative one and stops a stray relative path
  from silently creating a fresh, height-zero state.
- **Never put a password inline** in `tmkms.toml`. Use a separate
  `password_file` (Part B.5), itself `0600`, so the config can be read
  for debugging without exposing the credential.

## A.3 Choose the transport: Unix socket or firewalled TCP

This is where mutual authentication actually lives. Two valid layouts,
each with a *different* auth boundary — pick deliberately:

### Same host → Unix-domain socket (auth by filesystem)

If tmkms and gnoland run on the same machine, prefer a Unix socket.
gnoland creates it and `chmod`s it to `0600`, so only gnoland's user owns
it; place it in a directory only the tmkms and gnoland users can traverse:

```sh
sudo install -d -o gnoland -g tmkms -m 750 /run/gnoland
# gnoland will create /run/gnoland/privval.sock at 0600
```

> **Important:** over a Unix socket the `allowed_kms_pubkeys` allowlist is
> **not** the gate — gnoland cannot crypto-verify a UDS peer against it.
> *Filesystem permissions are the entire authentication boundary.* That
> is exactly why the socket directory and node-user separation in A.1–A.2
> matter: anyone who can open that socket becomes the signer. You still
> set the allowlist (gnoland requires a non-empty one), but on UDS it is
> belt-and-suspenders, not the lock.

### Separate hosts → TCP, firewalled (auth by cryptography)

For real stake, run the signer on its **own host**. Then the auth
boundary is cryptographic and works across the network:

- gnoland verifies tmkms via `allowed_kms_pubkeys` (the signer's pubkey).
- tmkms verifies gnoland via the **peer ID pinned in its `addr`**.

Both are mandatory. Then firewall the listen port so that **only the
signer host** can reach it — the crypto is the lock, the firewall keeps
strangers from even knocking (and from using your validator as a free
oracle to probe):

```sh
# allow only the signer host to reach the privval port; deny the rest
sudo ufw allow from <SIGNER_HOST_IP> to any port 26659 proto tcp
sudo ufw deny 26659/tcp
```

Bind gnoland's listener to the specific reachable interface, not a
wildcard, when you can.

The rest of this guide uses placeholders you can set to match your
choice. For a same-host UDS layout:

```sh
export PRIVVAL_LISTEN="unix:///run/gnoland/privval.sock"   # gnoland listens here
export TMKMS_ADDR="unix:///run/gnoland/privval.sock"       # tmkms dials here (no peer-ID on UDS)
```

For a two-host TCP layout (note the **mandatory** `<peer_id>@` prefix in
the tmkms address — filled in at B.3):

```sh
export PRIVVAL_LISTEN="tcp://<validator_interface_ip>:26659"
# TMKMS_ADDR set in B.5 once the peer ID is known: tcp://<peer_id>@<validator_ip>:26659
```

The two transports are independent of the signer backend, so the worked
examples below deliberately show both: **Part B (YubiHSM) uses TCP** (the
peer-ID-pinned, cross-host form), while **Parts C (softsign) and D
(Ledger) use the same-host Unix socket** (the simpler form). Use
whichever matches your real deployment regardless of which backend you
picked.

---

# Part B — Production validator with a YubiHSM (recommended)

A YubiHSM holds the consensus key inside tamper-resistant hardware and,
provisioned correctly, **never lets it out** — not even to its own
operator. That single property removes the entire "key stolen off disk /
exported silently" class of risk that softsign cannot. The catch is that
tmkms's *convenience* tooling around the YubiHSM is where the review found
its sharpest edges, so the rule for this section is: **provision the
device by hand, and never run the automated setup or test commands
against a production key.**

Install tmkms with the `yubihsm` provider (it is **not** in the default
build — you must enable the feature explicitly):

```sh
cargo install tmkms --version 0.15.0 --features yubihsm --locked
tmkms version   # → 0.15.0
```

## B.1 Provision the YubiHSM by hand — never `tmkms yubihsm setup`

> **Do not run `tmkms yubihsm setup`.** That command prints the recovery
> mnemonic, the wrap key, and the auth passwords to standard output
> (captured by every terminal logger and scrollback), and it provisions
> roles with key-export capability. Provision with `yubihsm-shell`
> instead, so the secrets never leave your control.

Work on an **offline** machine. The high-level policy you are encoding —
and which the rest follows from — is:

1. **Generate the consensus key inside the device** (or import it once and
   destroy the source). It must **not** carry the `exportable-under-wrap`
   capability. A non-exportable key cannot be wrapped out, so a stolen
   credential later cannot exfiltrate it.
2. **The auth key tmkms uses to log in gets `sign-eddsa` only.** No
   `export-wrapped`, no `sign-attestation-certificate`, no broad
   delegated capabilities. If that credential leaks, it can ask for
   signatures (which the firewall + state machine still gate) but cannot
   extract or clone the key.
3. **Turn the audit log on and force it.** The device ships with auditing
   off; with it forced on, the HSM records every operation and refuses to
   sign once its log buffer fills — so an attempted misuse leaves a trail
   and cannot run unbounded.
4. **Store the recovery seed offline, on paper, in a safe.** Never on the
   validator host.

The `<tmkms-auth-password>` below is a credential **you choose** — it is
what tmkms presents to log in to the device. Generate a long random one
(don't type a memorable string) and keep it only in the `0600`
`password_file` created further down; never commit it or paste it into
`tmkms.toml`:

```sh
TMKMS_AUTH_PASSWORD=$(openssl rand -base64 32)   # generate once; use it below and in the password_file
```

A representative `yubihsm-shell` session (verify the exact token spelling
against your firmware — capability and option names are the firm part).
Substitute `$TMKMS_AUTH_PASSWORD` for `<tmkms-auth-password>`:

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

The two things to never grant the signer's auth key: `export-wrapped` and
`sign-attestation-certificate`. The first turns a leaked credential into
key theft; the second lets it produce attestations you didn't intend.

Store that same password in a file, not the config:

```sh
printf '%s' "$TMKMS_AUTH_PASSWORD" | sudo tee /etc/tmkms/yubihsm-password >/dev/null
sudo chown tmkms:tmkms /etc/tmkms/yubihsm-password
sudo chmod 600 /etc/tmkms/yubihsm-password
unset TMKMS_AUTH_PASSWORD   # drop it from the shell once it's in the file
```

## B.2 Generate gnoland's node + listener identity

The consensus key lives in the HSM, but gnoland still needs its own node
identity (the peer ID that tmkms pins on TCP), and tmkms needs a
SecretConnection identity key. Generate the gnoland node secrets:

```sh
gnoland secrets init -data-dir /var/lib/gnoland/secrets
```

In tmkms mode gnoland's `priv_validator_key.json` is **not** used to sign
— tmkms signs. Only `node_key.json` (the peer ID) matters here.

**Peer ID in hex (TCP layouts).** tmkms pins your validator's peer ID to
make sure it signs for *your* node and not whoever answers on the port.
gnoland reports the identity in bech32 (`g1…`); tmkms wants the same 20
bytes in hex. Run the `nodeid-hex.go` helper from
[Prerequisites](#prerequisites--build-the-helper-tools):

```sh
export VALIDATOR_PEER_ID=$(go run ./nodeid-hex.go /var/lib/gnoland/secrets/node_key.json)
echo "$VALIDATOR_PEER_ID"   # e.g. 243cef06…dcac — pinned into tmkms.toml in B.5 (TCP)
```

**tmkms's SecretConnection identity (TCP layouts).** gnoland only accepts
the connection if this key's public half is in `allowed_kms_pubkeys`. Run
the `tmkms-identity-keygen.go` helper from
[Prerequisites](#prerequisites--build-the-helper-tools):

```sh
ALLOW=$(go run ./tmkms-identity-keygen.go /var/lib/tmkms/secrets/kms-identity.key)
sudo chown tmkms:tmkms /var/lib/tmkms/secrets/kms-identity.key
echo "$ALLOW"   # e.g. ed25519:4b6efade…b18a — copy for B.6
```

## B.3 Read the consensus pubkey out of the HSM and register the validator

Because the key never leaves the HSM, you can't push it into genesis —
you read the **public** half out of tmkms and register *that* as your
validator identity. Start tmkms once (with the `tmkms.toml` from B.4) and
it logs the device's consensus pubkey on connect:

```
[keyring:yubihsm] added consensus Ed25519 key: 2C854661478AA1CDC954D11ABA6ABB6DBF469572564C24C61ABFC0622A04D350
```

The hex string above is an **example** — copy the one *your* device logs.

Convert that hex pubkey to gno's `gpub1…` / `g1…` forms with this helper.
Write the file, then run it **from the gno repo root** (so the
`github.com/gnolang/gno/...` imports resolve):

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

That `gpub1…` / `g1…` pair is your validator's public identity. What you
do with it depends on whether you are joining a chain or bootstrapping
one:

**Joining an existing chain (the production case).** You do **not**
generate genesis — the chain already exists. Submit your `gpub1…` pubkey
and `g1…` address to that chain's validator-onboarding path (its staking
/ governance process, or the genesis coordinator if it hasn't launched
yet) to be added to the validator set. The pubkey→address relationship is
your integrity check: whoever registers you can confirm the address is
the hash of the pubkey. Keep the pair; you'll confirm gnoland signs with
this address in B.7. Then skip to [B.4](#b4-write-tmkmstoml).

**Bootstrapping your own chain (test / private networks).** Build genesis
explicitly with `gnogenesis`, seeding it with the HSM pubkey. Do **not**
use `gnoland start -lazy` here — it would mint a throwaway local key and
seed genesis with *that* instead of your HSM key:

```sh
gnogenesis generate -chain-id "$CHAIN_ID" -output-path /var/lib/gnoland/genesis.json
gnogenesis validator add \
  -genesis-path /var/lib/gnoland/genesis.json \
  -address  g1qmptf8uxdg6l0rh07jwvur0kk8my9vrdf5qtp4 \
  -pub-key  gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9z... \
  -name     hsm-validator \
  -power    10
```

`validator add` checks that the pubkey hashes to the address, so a
copy-paste slip is caught right here — the same integrity check the
onboarding path performs in the production case.

## B.4 Write `tmkms.toml`

Set the signer address to match the transport you picked in Part A.3 —
pick **one** (the TCP form embeds the mandatory peer-ID pin from B.2):

```sh
# TCP (separate hosts):
export TMKMS_ADDR="tcp://${VALIDATOR_PEER_ID}@<validator_ip>:26659"
# OR same host (Unix socket — no peer ID; socket perms are the boundary):
export TMKMS_ADDR="unix:///run/gnoland/privval.sock"
```

Write the config with a heredoc, then lock it down to `0600` owned by
`tmkms` (the password stays in the separate `password_file`, never
inline; all paths are absolute):

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

sudo chown tmkms:tmkms /etc/tmkms/tmkms.toml
sudo chmod 600 /etc/tmkms/tmkms.toml
```

The fields that carry security weight:

- **`addr` peer-ID pin (TCP).** The `<peer_id>@` prefix is the half of
  mutual auth that lets tmkms verify gnoland. **Omit it and tmkms will
  sign for whoever answers on that port** — an on-path attacker becomes a
  signing oracle — and it only logs `unverified validator peer ID!` while
  doing so. Always pin `$VALIDATOR_PEER_ID` from B.2. (On a `unix://`
  socket there is no peer ID; the socket permissions from Part A.3 are the
  boundary instead.)
- **`state_file` (absolute).** This is tmkms's authoritative double-sign
  (height/round/step) gate. It must be an absolute path on durable
  storage. **Never restore a stale copy of it** — rolling it back to an
  older height is the classic self-double-sign. See B.7 for the one
  durability caveat that remains.
- **`password_file`, not an inline `password`.** Keeps the credential out
  of the config and out of process listings.
- **`protocol_version = "v0.34"`.** gnoland speaks only the v0.34 privval
  dialect and refuses any other value. tmkms 0.15.0 prints a
  `deprecated protocol_version v0.34 (update to v0.38)!` warning — that is
  expected; stay on a tmkms release that still supports v0.34. Do **not**
  drop the field or bump it to v0.38 for gnoland.
- **No `state_hook`.** tmkms's hook subsystem for regenerating state has
  sharp edges (its stdout is never captured, so a `fail_closed` hook can
  silently kill startup). Write/seed the state file directly instead;
  don't wire a hook.

## B.5 Configure gnoland's tmkms listener

Create a default config, then set the four `tmkms_listener` fields. The
allowlist must be non-empty — gnoland refuses to enable the mode with an
empty one, because an empty allowlist means *accept any peer that
completes a handshake* (fail-open).

```sh
CFG=/var/lib/gnoland/config/config.toml
gnoland config init -config-path "$CFG"
```

> **Ordering gotcha.** `config set` validates the whole `tmkms_listener`
> block on every write, and a non-empty `listen_addr` is what *enables*
> that validation. Set `listen_addr` **last** — set it first and the
> write is rejected because `chain_id` is still empty, and it silently
> stays unset.

```sh
gnoland config set -config-path "$CFG" consensus.priv_validator.tmkms_listener.chain_id "$CHAIN_ID"
gnoland config set -config-path "$CFG" consensus.priv_validator.tmkms_listener.protocol_version "v0.34"
gnoland config set -config-path "$CFG" consensus.priv_validator.tmkms_listener.allowed_kms_pubkeys "$ALLOW"
# listen_addr LAST — this enables the mode:
gnoland config set -config-path "$CFG" consensus.priv_validator.tmkms_listener.listen_addr "$PRIVVAL_LISTEN"
```

Verify it stuck (`listen_addr` should be your `$PRIVVAL_LISTEN`, not
empty):

```sh
gnoland config get -config-path "$CFG" consensus.priv_validator.tmkms_listener
```

## B.6 Run both as services, in the right order

**Order matters.** gnoland blocks at startup for up to
`wait_for_connection_timeout` (default 60s) waiting for tmkms to dial in.
Start tmkms first (with `reconnect = true` it will retry until gnoland is
listening). Run each under its own user — for example with systemd units
(`User=tmkms` / `User=gnoland`), or manually:

```sh
# signer (as the tmkms user)
sudo -u tmkms tmkms start -c /etc/tmkms/tmkms.toml

# validator (as the gnoland user) — point -genesis at the network's
# genesis.json (the chain you're joining, or the one you built in B.3);
# either way NO -lazy, since you're not minting a local key.
sudo -u gnoland gnoland start \
  -data-dir /var/lib/gnoland \
  -genesis /var/lib/gnoland/genesis.json \
  -chainid "$CHAIN_ID"
```

## B.7 Verify it's really the HSM signing — and keep it that way

Within a few seconds the node should advance through heights.

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
  sudo systemctl stop tmkms     # or: sudo pkill -f 'tmkms start'
  # gnoland's committed height stops climbing...
  sudo systemctl start tmkms    # ...and resumes after tmkms reconnects
  ```

> **Benign log noise.** Even when signing works, gnoland periodically
> logs `SignerListener: accept failed … i/o timeout` and `already
> connected, dropping listen request`. The endpoint holds one live signer
> connection and these are idle re-accept attempts timing out. Watch the
> committed height, not these lines.

**Keep it safe in production:**

- **The one residual code-level risk — state-file durability.** tmkms
  writes its state file but, in 0.15.0, does not `fsync` it before
  signing the next block. A hard power loss in that window can roll the
  on-disk height *backward*, and signing again from the lower height is a
  double-sign. Mitigate operationally: run the signer host on reliable
  power (UPS), keep the state file on durable local storage (not a flaky
  network mount), and **never** restore an old state file after a crash —
  if you're unsure of its freshness, treat the validator as needing a
  fresh, carefully-reconciled state rather than a rollback. (A two-line
  upstream patch adding the `fsync` closes this; track it if you build
  tmkms yourself.)
- **Back up the state file** as part of every maintenance procedure — and
  understand a backup is for disaster recovery of the *latest* state, not
  a checkpoint to roll back to:

  ```sh
  sudo install -o tmkms -g tmkms -m 600 \
    /var/lib/tmkms/secrets/consensus_state.json \
    /var/lib/tmkms/secrets/consensus_state.json.bak
  ```
- **Drain the audit log.** With forced auditing on (B.1) the HSM stops
  signing once its audit buffer fills, so run a small auditor that reads
  and clears the log periodically (e.g. every 30s) and ships it off-box.
- **Never run `tmkms yubihsm test` against the production key.** It is an
  unbounded signing oracle that bypasses the double-sign state machine
  entirely. Test only against a throwaway key on a non-production device.
- **Don't re-run provisioning over a live key.** tmkms's key-write path
  has no "refuse if exists" guard, so a re-run of an init/keygen can
  silently overwrite. Provision once; back up before you ever touch it.

---

# Part C — Testnet/lab with softsign (key on disk)

softsign keeps the consensus key in a **file on the validator host**.
That is fine for understanding the wiring and for testnets, but the key
touches disk, so this path is **not for real stake** — use Part B for
that. Everything about the *connection* (Part A transport, allowlist,
peer-ID pin, `protocol_version`, verification) is identical; only key
custody and the genesis bootstrap differ.

You still apply Part A: dedicated `tmkms` user, `0600` secrets, locked
directories, and the same transport choice. The softsign-specific
hardening on top:

- **`0600` on `consensus.key` and the state file**, owned by `tmkms`. A
  readable softsign key *is* the validator's private key.
- **Swap exposure.** tmkms zeroizes the key on drop but does not `mlock`
  it, so it can be paged to disk. Disable swap on the signer host (or use
  encrypted swap) so the key never hits persistent storage in the clear:

  ```sh
  sudo swapoff -a                     # and comment swap out of /etc/fstab to persist
  ```
- **Don't re-run `secrets init` / keygen over an existing key file** —
  it can overwrite without warning. Back up first.

## C.1 Generate node secrets and export the consensus key

Install tmkms with the softsign feature (it is **not** in the default
build):

```sh
cargo install tmkms --version 0.15.0 --features softsign --locked
```

Generate the gnoland node secrets (this also creates the consensus key
that softsign will load and that genesis will register):

```sh
gnoland secrets init -data-dir ./gnoland-data/secrets
```

gnoland stores the ed25519 private key as base64 of the 64-byte
`seed‖pubkey`; tmkms softsign wants base64 of just the 32-byte **seed**.
Reslice and re-encode (no crypto — a byte slice — so plain `python3` is
fine):

```sh
mkdir -p ./tmkms/secrets
python3 - <<'PY' > ./tmkms/secrets/consensus.key
import json, base64
v = json.load(open("gnoland-data/secrets/priv_validator_key.json"))["priv_key"]["value"]
print(base64.b64encode(base64.b64decode(v)[:32]).decode(), end="")
PY
chmod 600 ./tmkms/secrets/consensus.key
```

This lab uses a **same-host Unix socket** (A.3) — the simplest transport
and the counterpart to Part B's TCP setup — so there is **no peer-ID pin
to generate**. You still need tmkms's SecretConnection identity key
(gnoland requires a non-empty allowlist even on a socket). Generate it
with the [Prerequisites](#prerequisites--build-the-helper-tools) helper:

```sh
ALLOW=$(go run ./tmkms-identity-keygen.go ./tmkms/secrets/kms-identity.key)
echo "$ALLOW"   # goes into gnoland's allowed_kms_pubkeys (C.3)
```

## C.2 tmkms.toml with the softsign provider

Same shape as B.4, with two lab changes: the provider block is
**softsign**, and the transport is a same-host **Unix socket** instead of
TCP (so no peer-ID pin — Part B already showed the TCP form). tmkms needs
**absolute** paths (A.2), so resolve the lab dirs first, then write the
file with a heredoc so the paths expand:

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

Because the consensus key exists on disk *and* is registered locally, you
can let `-lazy` build genesis from `priv_validator_key.json`.

First configure the listener. Create a default config, then set the four
`tmkms_listener` fields. The allowlist must be non-empty — gnoland
refuses to enable the mode with an empty one. On this Unix-socket layout
the allowlist is belt-and-suspenders — **filesystem permissions on the
socket are the real boundary** (A.3) — but gnoland still requires it
non-empty, so set it with `$ALLOW` from C.1.

```sh
export CHAIN_ID=gno-tmkms-test                 # must match C.2's [[chain]].id
export PRIVVAL_LISTEN="unix://$PRIVVAL_SOCK"   # gnoland listens on the socket from C.2
CFG=./gnoland-data/config/config.toml
gnoland config init -config-path "$CFG"
```

> **Ordering gotcha.** `config set` validates the whole `tmkms_listener`
> block on every write, and a non-empty `listen_addr` is what *enables*
> that validation. Set `listen_addr` **last** — set it first and the
> write is rejected because `chain_id` is still empty, and it silently
> stays unset.

```sh
gnoland config set -config-path "$CFG" consensus.priv_validator.tmkms_listener.chain_id "$CHAIN_ID"
gnoland config set -config-path "$CFG" consensus.priv_validator.tmkms_listener.protocol_version "v0.34"
gnoland config set -config-path "$CFG" consensus.priv_validator.tmkms_listener.allowed_kms_pubkeys "$ALLOW"
# listen_addr LAST — this enables the mode:
gnoland config set -config-path "$CFG" consensus.priv_validator.tmkms_listener.listen_addr "$PRIVVAL_LISTEN"
```

Verify it stuck (`listen_addr` should be your `$PRIVVAL_LISTEN`, not
empty):

```sh
gnoland config get -config-path "$CFG" consensus.priv_validator.tmkms_listener
```

Then start both:

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

- `-skip-genesis-sig-verification` is the standard dev-genesis flag (the
  default `-lazy` genesis contains example txs signed by test accounts);
  it is unrelated to tmkms.
- `-lazy` only generates *missing* files — it will not clobber your tmkms
  config or secrets — reads the local consensus pubkey to register the
  genesis validator, then boots using the tmkms listener for signing.

Verify exactly as in B.7.

---

# Part D — Advanced: Ledger hardware signer

A Ledger holds the consensus key **on the device** and never exports it —
like a YubiHSM, but it is a single signer with no failover, so it suits
advanced/low-stake setups rather than high-availability mainnet duty (for
HA use a YubiHSM or threshold signing such as Horcrux). The bootstrap is
the same "read the pubkey out, seed genesis from it" flow as Part B; only
the provider and the device handling change.

> **Residual hardware caveat.** In tmkms 0.15.0 the Ledger code path does
> not check the device's APDU return code on the pubkey and signature
> responses, so in principle a device/transport error could be accepted
> as if valid. In practice the `gnogenesis validator add` pubkey→address
> check (B.3) catches a wrong *pubkey* at setup time, and a malformed
> *signature* simply fails consensus rather than producing a bad valid
> one. Use a known-good device and verify the pubkey out-of-band.

## D.1 Prerequisites

Install tmkms with the Ledger provider. The cargo feature is named
**`ledger`** in 0.15.0 (older versions called it `ledgertm`, which is
why the config block below is still `[[providers.ledgertm]]`). tmkms
0.15.0 ships **no** default features, so plain `cargo install tmkms`
gives you neither Ledger nor softsign — pass them explicitly:

```sh
cargo install tmkms --version 0.15.0 --features ledger --locked
# or, to keep softsign too: --features ledger,softsign
```

You also need:

- A Ledger with the **"Tendermint Validator"** app installed (the
  dedicated ed25519 consensus app — *not* the regular Cosmos app),
  plugged in, unlocked, with that app **open**. tmkms cannot reach a
  locked device or a different app.

## D.2 Provider block and bootstrap

Use a `tmkms.toml` like B.4's, with two changes: the provider is
`ledgertm` (no key files — the key is on the device), and like Part C
this uses a same-host **Unix socket** (A.3) instead of TCP, so there is
no peer-ID pin. Create the socket directory as in A.3
(`sudo install -d -o gnoland -g tmkms -m 750 /run/gnoland`), then write
the config exactly as in B.4 (`sudo tee /etc/tmkms/tmkms.toml`) with:

```toml
[[chain]]
id = "${CHAIN_ID}"
key_format = { type = "hex" }
state_file = "/var/lib/tmkms/secrets/consensus_state.json"

[[providers.ledgertm]]
chain_ids = ["${CHAIN_ID}"]

[[validator]]
chain_id = "${CHAIN_ID}"
addr = "unix:///run/gnoland/privval.sock"   # same-host UDS; no peer-ID pin (A.3)
secret_key = "/var/lib/tmkms/secrets/kms-identity.key"
protocol_version = "v0.34"
reconnect = true
```

Only one `[[providers.ledgertm]]` is allowed and it always signs the
**consensus** key. Then follow B.2 (node + identity — on UDS the peer-ID
hex from `nodeid-hex.go` is unused, but you still need the kms-identity
key for the allowlist), B.3 (start tmkms, read `added consensus Ed25519
key`, `pkconv.go`, `gnogenesis`), B.5 (listener — set `listen_addr` to
the same `unix:///run/gnoland/privval.sock`), and B.6/B.7 (start
non-lazy, verify). gnoland should log `This node is a validator` with the
`gpub1…` from B.3 and climb.

> **Keep the device awake.** If the Ledger locks, sleeps, or the app is
> backgrounded, signing stops and the node stalls until you reopen the
> app. Plan for that on anything you care about.

---

# Part E — Prove it with the repo's automated test

A build-tagged Go test drives a **real tmkms binary** through the full
signing flow (softsign and Ledger backends), exercising the gnoland side
of every path above:

```sh
go test -tags=tmkms_integration -count=1 -v ./tm2/pkg/bft/privval/upstream/...
```

| Symptom | Cause | Fix |
|---|---|---|
| tmkms: `unknown field 'softsign', expected 'ledgertm'` | tmkms built without softsign | `cargo install tmkms --version 0.15.0 --features softsign --locked` |
| gnoland: `allowed_kms_pubkeys must not be empty` | allowlist unset, or `listen_addr` set before the other fields | Set the four fields with `listen_addr` **last** (B.5) |
| gnoland exits after ~60s waiting for connection | tmkms not running, can't reach `addr`, or pubkey not in allowlist | Start tmkms first; check `addr` vs `listen_addr`; confirm the hex in `allowed_kms_pubkeys` |
| `protocol_version` rejected at startup | a value other than `v0.34` | Use `v0.34` on both sides |
| gnoland panics at boot: `PubKey does not match Signer address` | default `-lazy` genesis has example txs (softsign path) | Add `-skip-genesis-sig-verification` (C.3) |
| Node never advances past height 1, no signing in tmkms log | `chain_id` mismatch across genesis / config.toml / tmkms.toml | Make all three identical and restart |
| Signatures produced but rejected by the chain | consensus key in the signer ≠ genesis validator key | Re-seed genesis from the *same* key the signer holds |
| tmkms: `unverified validator peer ID! (<hex>)` | peer-ID prefix missing from `addr` (TCP) | Pin `tcp://<hex>@host:port` with `$VALIDATOR_PEER_ID` (B.4) |
| tmkms (Ledger): startup `SigningError` / no pubkey | device locked, asleep, or wrong app | Unlock, open **Tendermint Validator** app, restart tmkms |

---

# Secure-setup checklist

Run down this list before putting real stake behind the validator. Each
item plainly states why it matters.

**Host & OS**
- [ ] tmkms runs as a dedicated, no-shell system user — *limits what a
  compromised signer process can reach.*
- [ ] Config, state, and key files are `0600`, owned by `tmkms` — *tmkms
  won't reject world-readable secrets; the OS must.*
- [ ] All paths in `tmkms.toml` are absolute — *stops a CWD-relative
  state file from silently becoming a fresh, height-zero one.*

**Transport & auth**
- [ ] Same host: Unix socket in a directory only tmkms+gnoland can reach —
  *on UDS, filesystem permissions are the only auth boundary.*
- [ ] Separate hosts: TCP with the peer-ID pinned in `addr` **and** the
  signer pubkey in `allowed_kms_pubkeys` — *both halves of mutual auth;
  without the pin tmkms signs for any impostor on the port.*
- [ ] Listen port firewalled to the signer host only — *keeps strangers
  from knocking or probing your signer.*
- [ ] `allowed_kms_pubkeys` is non-empty — *an empty list is fail-open.*

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
