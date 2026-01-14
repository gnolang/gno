# Stress Testing with Supernova

## What is Supernova?

Supernova is a stress-testing tool designed specifically for Gno TM2 networks.
It helps node operators and developers understand how their network behaves
under load by simulating realistic transaction patterns and measuring
performance metrics.

## Why Stress Test Your Network?

Before deploying a Gno network to production, you need to answer critical
questions:

- **What is the maximum throughput?** How many transactions per second (TPS)
  can your node handle before performance degrades?
- **How does gas consumption affect block production?** Understanding the
  relationship between gas limits and transaction throughput helps optimize
  block parameters.
- **Are there bottlenecks?** Stress testing reveals performance issues in
  consensus, storage, or network layers that may not appear during normal
  operation.
- **How do different workloads compare?** Deploying contracts versus calling
  methods have different resource requirements. Knowing this helps with
  capacity planning.

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
| REALM_DEPLOYMENT | Deploys a new realm per transaction | Testing heavy workloads (compilation, storage, state init) |
| PACKAGE_DEPLOYMENT | Deploys pure packages (stateless libraries) | Testing code storage overhead |
| REALM_CALL | Deploys one realm, then calls its methods | Simulating production workloads |

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

- **Low utilization (<50%)**: The network has spare capacity. Transaction
  volume is below what the network can handle.
- **High utilization (>80%)**: The network is near capacity. Consider
  increasing gas limits or optimizing transaction costs.
- **Variable utilization**: May indicate inconsistent transaction batching or
  network congestion patterns.

## When to Use Supernova

- **Before mainnet launch**: Validate your network can handle expected load
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

For installation and usage instructions, see the
[Supernova GitHub repository](https://github.com/gnolang/supernova).

For official benchmark results across different configurations, see the
[benchmark reports](https://github.com/gnolang/benchmarks/tree/main/reports/supernova).
