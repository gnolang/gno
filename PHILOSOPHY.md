# Philosophy

 * Simplicity of design.
 * The code is the spec.
 * Minimal dependencies - all dependencies must get audited.
 * Completeness - software projects that don't become
   complete are projects that are forever vulnerable.  One of
the primary goals of the Gno language and related works is to
become finished within a reasonable timeframe.

## Gno Philosophy

 * The interpreter serves as a spec for how the AST is meant to be interpreted.
 * The (virtual) machine is designed to interpret an immutable AST which matches the language.
 * The interpreter is meant to become independent of the host language, Go.
 * After the Gno interpreter can interpret itself, we will implement bytecode compilation.

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
