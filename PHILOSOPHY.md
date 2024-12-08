# Philosophy

 * Simplicity of design - there should be one obvious way to do it.
 * The code is the spec.
 * Readability is paramount - beautiful is better than fast.
 * Minimal code - keep total footprint small.
 * Minimal dependencies - all dependencies must get audited, and become part of the repo.
 * Modular dependencies - wherever reasonable, make components modular.
 * Finished - software projects that don't become finished are projects that
   are forever vulnerable. One of the primary goals of the Gno language and
   related works is to become finished within a reasonable timeframe.
 * Maintainable, debuggable, and future-proof codebase.

## Gno Philosophy

 * The interpreter serves as a spec for how the AST is meant to be interpreted.
 * The (virtual) machine is designed to interpret an immutable AST which matches the language.
 * The interpreter is meant to become independent of the host language, Go.
 * After the Gno interpreter can interpret itself, we will implement bytecode compilation.

## Tendermint Philosophy

 * Each node can run on a commodity machine. Corollarily, for scaling we focus on sharding & forms of IBC.

## Performance Philosophy

* Correct, debuggable software is more important than extreme performance.
* Multicore concurrency makes Tendermint within the range of theoretical performance.
* Go is chosen for faster development of modular components, not for maximum speed.
* Real bottleneck is in the application layer, not in supporting large validator sets.
* Focus on feature completeness, debuggability, and maintainability over extreme performance.

## CLI Philosophy

 * No envs.
 * No short flags, with the following exceptions:
   * `-h` for showing help
   * `-v` for being verbose
   * mimicking the short flags of Go commands
   * after software maturity
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
