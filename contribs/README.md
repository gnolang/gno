# Gno Contribs

This directory contains additional tools to enhance your Gno experience. Tools
can be simple wrappers for `gno`, `gnoland`, and `gnokey`, or complete
applications.
They may be Go binaries with their own `go.mod` files, shell scripts.
We currently only accept Go and shell scripts, but we're open to discussing
other languages.

Tools found here are experimental and may be either sunsetted or promoted to official
tools.

## Contributing Guidelines

If you'd like to contribute a tool, please follow these guidelines:

1. **Naming**: Use a clear name starting with `gno`, `gnoland`, or `gnokey`,
   followed by a descriptive word. This helps users understand the tool's
   purpose.

2. **User-Friendly**: Follow the style of other Gno tools for a consistent
   experience.

3. **Maintenance**: Tools are experimental and require maintenance to stay
   compatible. If a tool becomes incompatible and the core team cannot maintain
   it, we may disable it until you update it, or sunset it completely. In that
   case, interested users can fork it to an external repository.

4. **CI**: Include a Makefile with 'install', 'lint' and 'test' targets to help
   identify compatibility issues early.