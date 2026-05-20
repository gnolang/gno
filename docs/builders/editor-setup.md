# Editor Setup

This guide helps you configure your editor for working with `.gno` files —
autocompletion, go-to-definition, diagnostics, and formatting.

## Language server

[gnopls](https://github.com/gnoverse/gnopls) is the Gno language server - a
fork of [gopls](https://github.com/golang/tools/tree/master/gopls) adapted
for Gno. It works with any editor that supports the
[Language Server Protocol](https://microsoft.github.io/language-server-protocol/) (LSP).

Install the [Gno toolchain](install.md) first, then follow the
[gnopls README](https://github.com/gnoverse/gnopls#readme) for editor-specific
setup.

- **VS Code** — install the [Gno for VS Code](https://marketplace.visualstudio.com/items?itemName=Gnoverse.gnolang)
  extension; it bundles `gnopls`, so you can skip the manual install.

## Verify your setup

`gnopls version` only proves the binary runs. To confirm your editor is
actually talking to it, open any `.gno` file and check:

- **Completion** — start typing `pri`; `println` should appear in the
  suggestion list.
- **Hover** — hover a symbol like `println`; you should see its signature
  and docs.
- **Diagnostics** — break something on purpose (e.g. a typo in an import
  path); the editor should underline it within a second or two.

If none of that happens, the editor isn't connected to `gnopls` — check
the language server logs in your editor.

## Next steps

### Format on save

`gno fmt` is wired through gnopls, but format-on-save isn't automatic.
Enable it in your editor's settings — for example, VS Code's
`editor.formatOnSave`, or Neovim's `BufWritePre` autocmd.

### Hot reload with gnodev

`gnodev` runs a local Gno.land node that hot-reloads on save: edit a
`.gno` file, save it, and refresh `gnoweb` to see the result. See
[Local development with `gnodev`](../resources/gnodev.md).

## Contributing

`gnopls` is under active development and tracks `gopls` with some lag —
if something doesn't work as expected, please open an issue or PR on the
[gnopls repository](https://github.com/gnoverse/gnopls).
