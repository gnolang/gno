# ADR-prxxxx: Config loader must honor explicit zero values

## Status

Proposed

## Context

`config.LoadConfig` (used by `gnoland start`) and `LoadOrMakeConfigWithOptions`
loaded `config.toml` in two steps:

1. `LoadConfigFile` decoded the TOML document into a **fresh, zero-valued**
   `Config` struct.
2. `mergo.Merge(loadedCfg, DefaultConfig())` filled any zero-valued field of the
   loaded config from the defaults, intending to "supply defaults for keys the
   file omits".

mergo's default mode cannot distinguish "absent from the file" from "explicitly
set to the Go zero value". Every field decoded from an absent key and every
field the operator explicitly set to a zero value (`false`, `0`, `""`, empty
slice) look identical after step 1 — both are the zero value. Step 2 then
overwrites both from the defaults.

The consequence: any config field whose default is non-zero cannot be set to its
zero value via `config.toml`. The merge silently reverts it to the default, with
no error and no warning, even though the written file shows the intended value.

Concrete confirmed cases:

- **`consensus.create_empty_blocks = false`** — default `true`
  (`tm2/pkg/bft/consensus/config/config.go`). Silently reverted to `true`.
- **`rpc.cors_allowed_methods = []`** — default is a non-empty 4-element list
  (`tm2/pkg/bft/rpc/config/config.go`). An explicit empty list is silently
  refilled with the defaults.
- Any other zero-valued-but-non-zero-default field across `RPC`, `P2P`,
  `Mempool`, `Consensus`, `TxEventStore`, `Telemetry`, `Application`.

Non-zero values in the file (e.g. a non-empty `persistent_peers` string, a
custom `send_rate`) were honored, because mergo leaves a non-zero destination
field untouched. That asymmetry — non-zero values survive, `false`/`0`/`""`/`[]`
revert — is the mergo zero-value signature.

## Decision

Decode the TOML document **on top of a `DefaultConfig()`-initialized struct**,
and remove the `mergo.Merge` step entirely.

A TOML decoder distinguishes absent keys from present ones: decoding into a
pre-populated struct sets the fields present in the document (including explicit
`false`/`0`/`""`) and leaves absent fields at their existing (default) value.
This is exactly the semantics the mergo step was trying, and failing, to
approximate.

- `LoadConfigFile` now starts from `DefaultConfig()` and decodes the file over
  it, via a new unexported `loadConfigFile(path string, cfg *Config) error` that
  decodes into a caller-provided `*Config`.
- `LoadConfig` calls `LoadConfigFile` and drops the mergo merge.
- `LoadOrMakeConfigWithOptions` applies `opts` to `DefaultConfig()` first, then
  decodes the file over the result with `loadConfigFile`, and drops the mergo
  merge. Precedence is therefore **file > options > defaults**: a key present in
  the file wins over an option; an option is preserved for keys the file omits.
- The `dario.cat/mergo` dependency is no longer used in this package and is
  removed from `go.mod`.

The TOML library in use is `github.com/pelletier/go-toml v1.9.5`. Its
decode-over-a-pre-populated-struct behavior was verified before finalizing:
present scalar keys (including explicit `false`) overwrite the field, absent
keys leave the field untouched, and a slice present in the document **replaces**
the default slice rather than appending to it (so round-tripping the non-empty
default `rpc.cors_allowed_methods` does not duplicate entries).

## Alternatives considered

- **`mergo.Merge` with `WithOverride`**: makes the defaults override the file —
  backwards, strictly worse.
- **Sentinel/pointer fields to detect "unset"**: change every bool/int/string in
  every sub-config to a pointer and treat `nil` as unset. Large, invasive,
  churns the whole config surface and the CLI reflection in
  `gnoland config get/set`. The decode-over-defaults approach gets the same
  result with no type changes.
- **Per-field CLI flag overrides**: could let specific settings be overridden
  from the command line, but only addresses the individual fields wired up and
  does not fix the general zero-value class. Out of scope here.

## Consequences

- Operators can now set any config field to its zero value via `config.toml`
  (e.g. `consensus.create_empty_blocks = false`, `rpc.cors_allowed_methods = []`)
  and have it honored at runtime.
- `LoadConfigFile` returns a config with defaults applied for absent keys rather
  than Go zero values. This also affects `gnoland config get`, which now reports
  the value the node would actually use for a key omitted from the file (the
  default) instead of the Go zero value. Since `gnoland config init` writes a
  complete config, this only surfaces for hand-edited partial files, and the new
  behavior is the more correct one.
- `gnoland config set` on a partial config file materializes absent keys to
  their defaults when it rewrites the file. This produces a complete, valid
  config and does not change any value the node was already using.
- One dependency (`dario.cat/mergo`) removed from `go.mod`.

## Tests

`tm2/pkg/bft/config/config_test.go`:

- `TestConfig_LoadConfig`: an explicit `false` for boolean fields is honored; an
  explicit empty array (`rpc.cors_allowed_methods = []`) disables the non-empty
  default instead of reverting to it; keys/sections absent from the file keep
  their defaults; non-zero values (including a slice) load unchanged and are not
  doubled.
- `TestConfig_LoadOrMakeConfigWithOptions`: a file value takes precedence over an
  option for the same key; an option is preserved when the key is absent from the
  file.

The explicit-zero-value tests fail against the previous mergo-based
implementation and pass after the change.
