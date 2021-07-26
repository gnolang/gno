## General Philosophy

To avoid subversion, create an objective bar of competency and character.
Do not create a "team" to jointly create something when it can be done alone.
Require technical contributions that are well designed and encapsulated.
Let the light shine through, and let the chips fall where they may,
for the basis of accountability is the light, is transparency.

## Technical Philosophy

The code is meant to become the spec.
The interpreter serves as a spec for how the AST is meant to be interpreted.
The (virtual) machine is designed to interpret an immutable AST which matches the language.
The interpreter is meant to become independent of the host language, Go.
After the Gno interpreter can interpret itself, we will implement bytecode compilation.

### Command-line Philosophy

 * No envs.
 * No short flags.
 * No /bin/ calls.
 * No process forks.
 * Struct-based command options.
