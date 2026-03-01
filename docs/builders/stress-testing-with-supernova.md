# Stress Testing with Supernova

## What is Supernova?

[Supernova](https://github.com/gnolang/supernova) is a stress-testing
tool designed specifically for Gno TM2 networks. It helps node operators
and developers understand how their network behaves under load by
simulating realistic transaction patterns and measuring performance
metrics.

## Why Stress Test Your Network?

Stress testing answers critical questions before production deployment:

| Question | Why It Matters |
|----------|----------------|
| What's my maximum TPS? | Know when performance will degrade under load |
| How do gas limits affect throughput? | Optimize block parameters |
| Where are the bottlenecks? | Find issues in consensus, storage, or network |
| How do workloads differ? | Plan capacity for deployments vs calls |

## How Supernova Works

Supernova operates in three phases:

### 1. Account Setup

Supernova derives multiple subaccounts from a single mnemonic. This simulates
realistic conditions where transactions come from many different addresses,
testing the network's ability to handle concurrent account state updates.

The tool automatically distributes funds from the main account (index 0) to
all subaccounts before the test begins.

### 2. Transaction Generation

Based on the selected mode, supernova constructs and signs transactions.

## Stress Testing Modes

| Mode | What it Does | Best For |
|------|--------------|----------|
| REALM_DEPLOYMENT | Deploys a new realm per tx | Heavy workloads |
| PACKAGE_DEPLOYMENT | Deploys pure packages | Code storage |
| REALM_CALL | Deploys realm, calls methods | Production |

For most production scenarios, REALM_CALL provides the most relevant metrics
since it simulates typical user interactions.

### 3. Result Collection

After broadcasting transactions, supernova monitors the blockchain to collect
metrics like TPS, block utilization, and gas consumption.

## Understanding the Results

### TPS (Transactions Per Second)

TPS reflects real-world throughput, accounting for:
- Transaction propagation time
- Block production intervals
- Consensus overhead

A higher TPS indicates better network performance, but the optimal value
depends on your hardware, network configuration, and block parameters.

### Block Utilization

Block utilization reveals how efficiently blocks are being filled:

- **Low utilization (\<50%)**: The network has spare capacity. Transaction
  volume is below what the network can handle.
- **High utilization (\>80%)**: The network is near capacity. Consider
  increasing gas limits or optimizing transaction costs.
- **Variable utilization**: May indicate inconsistent transaction batching or
  network congestion patterns.

## When to Use Supernova

- **Before deployment**: Validate your network can handle expected load
- **After configuration changes**: Verify block gas limits, timing parameters
- **During capacity planning**: Determine hardware requirements for target TPS
- **Comparing node implementations**: Benchmark different setups objectively

## Integration with Monitoring

For deeper insights, run supernova against a node with
[OpenTelemetry enabled](../../misc/telemetry/README.md). This allows you to
correlate supernova's transaction metrics with internal node metrics like:

- Memory and CPU usage during load
- Consensus round timing
- Storage I/O patterns
- Network message latency

## Getting Started

### Prerequisites

- Go 1.19 or higher
- A running Gno node (e.g., via `gnodev` or `gnoland start`)
- A funded mnemonic (the first derived address needs funds for distribution)

### Installation

```bash
# Clone the repository
git clone https://github.com/gnolang/supernova.git
cd supernova

# Build the binary
make build

# The binary will be at ./build/supernova
```

### Basic Usage

```bash
./build/supernova \
  -url http://localhost:26657 \
  -mnemonic "source bonus chronic canvas draft south burst lottery \
vacant surface solve popular case indicate oppose farm nothing bullet \
exhibit title speed wink action roast" \
  -sub-accounts 5 \
  -transactions 100 \
  -mode REALM_CALL \
  -output results.json
```

You can check the results in `results.json`. For production-grade
testing, increase `-sub-accounts` (50-100) and `-transactions` (5000+).

### CLI Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-url` | (required) | JSON-RPC URL of the Gno node |
| `-mnemonic` | (required) | Mnemonic for deriving accounts |
| `-sub-accounts` | 1 | Number of accounts sending transactions |
| `-transactions` | 10 | Total transactions to send |
| `-mode` | REALM_DEPLOYMENT | Transaction mode (see Modes section) |
| `-batch` | 100 | Batch size for JSON-RPC calls |
| `-chain-id` | dev | Chain ID of the network |
| `-output` | (none) | Path to save results JSON |

### Resources

- [Supernova GitHub repository](https://github.com/gnolang/supernova)
- [Benchmark reports](
  https://github.com/gnolang/benchmarks/tree/main/reports/supernova)
