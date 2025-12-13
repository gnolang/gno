# `gnoland`

`gnoland` is the production binary powering the gno.land chain. You might want to run it if you're debugging low-level behavior or building advanced tooling like indexers.

For **local development**, we recommend using [gnodev](../../../contribs/gnodev) — a developer-optimized node that makes writing and testing Gno contracts much easier.

> Note: The `gnoland` binary is **specific to the gno.land chain**. Other chains in the Gno ecosystem will use different binaries tailored to their own configurations and goals.

## Getting Started

### Install `gnoland`

```bash
git clone git@github.com:gnolang/gno.git
cd gno/gno.land
make install.gnoland
```

### Quick Start (Development Mode)

For quick local testing, you can start a node with default settings:

```bash
gnoland start -lazy -skip-genesis-sig-verification
```

This command:
- `-lazy`: Delays transaction execution for better development experience
- `-skip-genesis-sig-verification`: Skips signature verification during genesis (development only)

### Full Setup (Production Mode)

For a production-like setup, initialize the configuration first:

1. **Initialize configuration:**
   ```bash
   gnoland config init
   ```
   This creates the necessary configuration files in your data directory.

2. **Initialize secrets (validator keys):**
   ```bash
   gnoland secrets init
   ```

3. **Start the node:**
   ```bash
   gnoland start
   ```

### Configuration

The node configuration is stored in your data directory (default: `$HOME/.gnoland`). You can:

- View configuration: `gnoland config get`
- Update configuration: `gnoland config set <key> <value>`
- Manage validator keys: `gnoland secrets get/verify`

### Interacting with the Node

Once running, you can interact with it using:
- [gnokey](../gnokey) – CLI wallet & tool
- [gnoweb](../gnoweb) – Web-based interface
