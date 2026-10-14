# ADR: gnoweb re-reads static alias files when they change

## Context

The gno.land homepage is a static alias (`--aliases=/=static:home.md`).
`misc/deployments/home-alias/gnoalias.sh` rewrites `home.md`, now hourly
with the latest blog posts. gnoweb read the file once, at startup, so the
script restarted gnoweb through the Docker socket mounted in its container:
a short outage on each update, a container with root-equivalent access to
its host, and a default command that restarted every `gnoweb:master`
container, stopped ones included.

## Decision

A static alias built from a file (`NewStaticFileAlias`, used by
`parseAliases`) keeps its path. When the page is served
(`prepareIndexBodyView`), gnoweb stats the file and re-reads it if its size
or modification time changed. A file modified less than two seconds before
it was read is read again on the next request, so a write within the same
timestamp tick, or an in-place write caught halfway, is not cached as
current. If the file cannot be read, the last content read keeps being
served. A missing file still fails at startup.

`gnoalias.sh` now only writes `home.md` (a new file renamed over the old
one, so a read never sees half a file). Its reload command is optional, for
a gnoweb older than this change; the compose file no longer installs Docker
or mounts its socket.

## Alternatives considered

- A file watcher (fsnotify): a new dependency, and unreliable across bind
  mounts and the symlink swaps of Kubernetes ConfigMaps. A stat per request
  costs microseconds next to the markdown render, and follows symlinks.
- A minimum interval between checks: not worth it at this cost; it would
  matter on a slow network filesystem.
- A reload signal (SIGHUP): still needs a process that can signal gnoweb,
  which is what the Docker socket was for.

## Consequences

- An operator can update a static page by replacing its file; no restart.
- The script and gnoweb must share the folder, not the file alone: a
  single-file bind mount keeps the old file after a rename (README).
- Until the deployed gnoweb includes this change, a reload command must
  still be passed to the script.
