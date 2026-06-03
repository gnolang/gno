# showcase

A documentation kitchen-sink package. It exists to exercise gnoweb's package
**source / overview** page: every exported symbol kind is declared here, so the
overview renders all of its sections and prefixes each symbol with its kind
glyph.

## Usage

```go
import "gno.land/p/demo/showcase"

it := showcase.New("Widget", 1500)
it.Publish()
```

## What it demonstrates

- **Constants** — `Version`, plus a `Status*` block.
- **Variables** — `DefaultTags`, `MaxTitleLen`.
- **Types** — struct, interface, slice, map, pointer, func and a defined
  primitive: one symbol per kind glyph.
- **Functions & methods** — package functions, and methods grouped under their
  receiver type.

See the package doc comment for a rendered example.
