# PR6440: A configurable mode for the tmkms_listener Unix socket

## Status

Proposed

## Context

With `listen_addr = "unix://…"` under `[consensus.priv_validator.tmkms_listener]`, gnoland set the socket to a fixed `0600` right after `net.Listen`. Only gnoland's user could connect, so the signer had to run as that user, and a compromise of gnoland then also reached what the signer holds: the HSM password (or a softsign key) and tmkms's double-sign state. The alternative was a root step after every node start that waited for the socket and gave it the signer's group and `0660`.

The group needs no gnoland change: a socket created in a setgid directory takes the directory's group (inode(7)), even when the creating user is not in that group. The fixed mode is what kept the signer's group out.

Two smaller defects sat in the same code:

- Until the chmod, the socket had `0777` minus the umask. With a `022` umask that is `0755`, which no other user can connect to (connecting needs write permission). With `002` or `000`, the group or every local user could connect in that window, and the first client to connect becomes the signer: `SignerClient.Init` caches the key it reports.
- A failed chmod only logged a warning and kept the umask's mode, although on a Unix listener the access control is the socket's mode and its directory's permissions (`allowed_kms_pubkeys` is ignored there).

Measured with a Go program listening as UID 1001 and clients as other UIDs (Linux, `golang` container):

| Setup | Socket | Other UID |
|---|---|---|
| umask `022` | `0755` | Permission denied |
| umask `000` | `0777` | connects |
| umask `000`, `fchmod(fd, 0600)` before `bind` | `0600` | Permission denied |
| setgid directory `2710` (owner the node, group the signer's, node not in it), then `0660` | created `0755` with the signer's group, then `0660` | the signer's UID connects, others are denied |

On macOS, `fchmod` on an unbound socket fails with `EINVAL`.

## Decision

1. A `socket_mode` key under `[consensus.priv_validator.tmkms_listener]`: an octal string, default `"0600"` (the previous fixed mode). One helper parses it with `strconv.ParseUint(s, 8, 32)` for both `ValidateBasic` and the listener. The mode must be within `0770`. It is checked for `unix://` only and ignored for `tcp://`.
2. On Linux, the socket is `fchmod`ed to the mode before `bind(2)`, through `net.ListenConfig.Control`, so `bind` creates the file with the mode minus the umask: never more open than the mode. Then, on every OS, `os.Chmod` sets the exact mode, since the umask may have removed bits the operator asked for.
3. If the mode cannot be set, the listener is closed, which removes the socket, and the start fails.
4. `ValidateBasic` refuses an empty `unix://` address and one starting with `@`: on Linux both name an abstract socket, which has no file, so no mode can protect it.

## Alternatives considered

- **A loopback `tcp://` listener with `allowed_kms_pubkeys`:** it already lets the signer run under its own user, needs no change, and stays available. UDS is kept usable that way because `gno.land/cmd/gnoland/TMKMS.md` recommends it when the signer is on the same host, and because over TCP every local user reaches the SecretConnection handshake, which leaves the signer's identity key as the only boundary; on UDS the kernel refuses `connect(2)` to every user the filesystem permissions leave out.
- **An integer key:** go-toml v1.9.5 writes `0o660` as `432` and rejects `0660`, JSON rejects a leading zero, and `gnoland config set` has no case for an `os.FileMode` field. An octal string reads and edits as the mode it is.
- **Base 0 parsing:** it would accept `0o660` and `0660`, but also `0x…` and decimal, where `660` would mean `01224`.
- **Any mode up to `07777`:** an "other" bit lets every local user connect and become the signer, and setuid, setgid and sticky have no use on a socket.
- **A `socket_group` key:** a non-root process can only give a file a group it belongs to (chown(2)), so gnoland's user would have to join the signer's group, which also grants it whatever else that group can open. The setgid directory gives the socket that group without it.
- **Rejecting or warning on a non-default `socket_mode` with `tcp://`:** every generated `config.toml` carries the key, and an operator moving a listener from `unix://` to `tcp://` would have to delete a key that no longer applies, or read a warning about it on every start. It is ignored, as the field's comment says.
- **Closing the window outside Linux:** changing the process umask around `Listen` affects every goroutine creating a file at that moment, and binding under a private name then renaming into place leaves the final path behind on `Close`, since Go unlinks the name it bound. Outside Linux the socket keeps `0777` minus the umask between `bind` and the chmod.
- **Leaving the failed-chmod path untested:** that path is the security change, since a node must not run on a socket whose mode it could not set. A test swaps a package variable, `chmodSocket`, for a failing one and checks that the start fails and no socket is left. A real chmod failure on the process's own socket cannot be produced in a test. The test is not parallel, so the swap cannot race the parallel tests that build a listener.

## Consequences

- A `config.toml` without the key keeps the default `"0600"`, the previous mode. `gnoland config init` writes the key; `gnoland config get/set` handle it as a string.
- Stricter than before: a mode that cannot be set stops the start instead of logging a warning, on Linux the socket never passes through the umask's mode, and an abstract `unix://` address, which used to start the node on a socket no permission protects, is refused when the config is loaded or set.
- A signer under its own user needs a socket directory owned by gnoland's user and the signer's group with mode `2710`, and `socket_mode = "0660"`; `gno.land/cmd/gnoland/TMKMS.md` documents it.
- Outside Linux, the window between `bind` and the chmod remains.
- `TestBindUnixSocketCreatesTheSocketWithItsMode` (Linux only) changes the process umask, so it is not parallel either; CI runs the Go tests on Linux.
