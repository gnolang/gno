# Gno Playground

Web-based IDE for writing, testing, and deploying Gno smart contracts.
Live at [play.gno.land](https://play.gno.land).

## Quick Start

```bash
# Development (builds WASM from local gnovm)
make dev

# Production build
make build

# Build using a release tag
make build-release TAG=chain/gnoland1.0
```

## Prerequisites

- Go 1.22+
- Node.js 20+
- pnpm 10+

## Architecture

```
contribs/playground/
├── app/                   # Main React application
├── packages/
│   ├── wasm/              # GnoVM WASM runtime
│   ├── core/              # Chain config, editor, services
│   ├── react/             # Shared React components
│   ├── pkg/               # Filesystem and git utilities
│   ├── codemirror-lsp/    # CodeMirror + LSP integration
│   └── vite-plugins/      # Custom Vite build plugins
├── tools/data/            # Build artifacts (gitignored)
│   ├── gno/root.zip       # Gno stdlibs + examples
│   └── wasm/gno.wasm      # GnoVM WASM binary
└── Makefile
```

## Build Modes

### Current (development)
Builds WASM from the local `../../gnovm` source tree. Use this to test
the playground against the current state of the repository.

### Release
Downloads pre-built WASM and root.zip from a GitHub release tag.
Used for production deployments.

```bash
make build-release TAG=chain/gnoland1.0
```

## Deployment

See `misc/deployments/playground/` for deployment configurations.
