# Validator scenario startup and readiness

## Context

The validator scenario harness runs Docker containers or native processes.
Docker can fail while publishing an ephemeral host port. Local runs previously
probed for listeners without reserving ports. Concurrent runs could select the
same port, including when a validator was temporarily stopped for a reset.
An HTTP success from another node could then mark a failed start as successful.

The generated Compose file was also rewritten after initial startup with the
resolved host ports. Removing a container on retry consequently requested the
same fixed port. Relative binary output paths were interpreted after different
build recipes changed directories.

## Decision

- Keep local port selection in Bash, avoiding existing listeners and ports
  already selected for nodes within the same scenario. The default usage is
  one local scenario at a time, without a helper binary or allocation process.
  Parallel local runs must use disjoint port ranges and work directories.
- Require the expected node ID in `/status`, and check the owned local process
  before accepting readiness. Bound each readiness HTTP request to two seconds.
- Keep ephemeral mappings in the Compose specification and record resolved
  ports in memory and inventory, refreshing inventory after a node restart.
  Start sentries, then validators, through the
  same per-service retry path. Bind conflicts retry at most three times;
  unrelated failures abort immediately. Explicit port hints remain fixed.
- Resolve `BIN_DIR` before changing build directories.

## Alternatives considered

- A shared port allocator can coordinate concurrent runs and protect stopped
  nodes' ports, but adds a mandatory binary and process to single-scenario
  runs. That coordination is outside the default execution scope.
- Holding TCP listeners would reserve ports against all programs, but requires
  listener-descriptor handoff support in gnoland and the signer. That would
  expand a harness fix into node startup changes.
- Project-wide Docker retries remove already-running sentries and signers.
  Per-service retries preserve them and avoid the former pipeline's interaction
  between `pipefail` and a failed `compose up` command.

## Consequences and verification

No VM, consensus, gas, database, or chain-replay behavior changes. Local runs
keep the same four binaries. The connect probe does not reserve ports or detect
a socket bound before listen. Conflicts fail startup instead of being treated
as a successful start. Automatic coordination between concurrent scenarios
with overlapping port ranges is intentionally unsupported.

Go tests execute the shell helpers against real HTTP endpoints and controlled
Docker/compiler boundaries. They cover wrong or missing RPC identity, dead
processes, local port selection without a helper, ephemeral Compose mappings
and refreshed inventory, bounded retries, preservation of other services, and
relative build output directories. The misc CI matrix runs these tests; real
local scenarios exercise consensus, reset/restart, and signers.
