# Philosophy

 * Simplicity of design.
 * The code is the spec.
 * Minimal code - keep total footprint small.
 * Minimal dependencies - all dependencies must get audited, and become part of the repo.
 * Modular dependencies - whereever reasonable, make components modular.
 * Finished - software projects that don't become finished are projects that
   are forever vulnerable. One of the primary goals of the Gno language and
   related works is to become finished within a reasonable timeframe.

## Gno Philosophy

 * The interpreter serves as a spec for how the AST is meant to be interpreted.
 * The (virtual) machine is designed to interpret an immutable AST which matches the language.
 * The interpreter is meant to become independent of the host language, Go.
 * After the Gno interpreter can interpret itself, we will implement bytecode compilation.

## Tendermint Philosophy

 * Each node can run on a commodity machine. Corollarily, for scaling we focus on sharding & forms of IBC.

## Cli Philosophy

 * No envs.
 * No short flags.
 * No /bin/ calls.
 * No process forks.
 * Struct-based command options.

## Token Philosophy

 * Single base token.
 * Deflationary is sufficient when tx fees are imminent.
 * Int64 is sufficiently large to handle human numbers; for everything else, use denom conversions.

## Spiritual Philosophy

 * Truth is revealed in the open light.
 * Everybody has the potential to see the light.
 * Those who see the light are moved to expand it.
