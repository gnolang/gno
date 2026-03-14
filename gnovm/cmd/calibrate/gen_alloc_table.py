#!/usr/bin/env python3
"""
Generate allocGasTable for GnoVM from Go allocation benchmarks.

Model: 6 exact table points (1B-32B) + power-law fit (64B onwards).
The power law ns = C * size^α is a straight line in log-log space.
Table entries at powers of 2 are populated from these. At runtime, allocGas()
uses bits.Len64 + linear interpolation (O(1), ~1.5ns).

NOTE: This script outputs values in milligas. Divide by 1000 for gas units
when updating allocGasTable in alloc.go.

Usage:
    # Run benchmarks (on the target deployment hardware):
    go test -bench=BenchmarkAlloc -benchmem -count=10 -timeout=30m . > bench_output.txt

    # Generate table and plot:
    python3 gen_alloc_table.py bench_output.txt

    # With custom cpuBaseNs (ns per gas unit, from benchops):
    python3 gen_alloc_table.py bench_output.txt --cpu-base-ns 10.0
"""
import argparse
import re
import sys
import numpy as np
from collections import defaultdict

def parse_benchmarks(path):
    """Parse Go benchmark output, return {size: [ns_values]}."""
    data = defaultdict(list)
    with open(path) as f:
        for line in f:
            m = re.match(r'BenchmarkAlloc_(\d+)-\d+\s+\d+\s+([\d.]+)\s+ns/op', line)
            if m:
                data[int(m.group(1))].append(float(m.group(2)))
    return data

def build_table(data):
    """Build 32-entry ns table: 6 exact points + power-law for the rest."""
    medians = {s: np.median(vs) for s, vs in data.items()}

    # First 6 entries: exact from benchmarks (1B, 2B, 4B, 8B, 16B, 32B)
    exact_sizes = [1, 2, 4, 8, 16, 32]
    exact_ns = []
    for s in exact_sizes:
        if s not in medians:
            print(f"ERROR: no benchmark data for {s}B", file=sys.stderr)
            sys.exit(1)
        exact_ns.append(medians[s])

    # Enforce monotonic
    for i in range(1, len(exact_ns)):
        if exact_ns[i] < exact_ns[i-1]:
            exact_ns[i] = exact_ns[i-1]

    # Power-law fit to all data from 64B onwards: log2(ns) = log2(C) + α*log2(size)
    fit_sizes = sorted([s for s in medians if s >= 64])
    if len(fit_sizes) < 3:
        print("ERROR: need at least 3 data points >= 64B for power-law fit", file=sys.stderr)
        sys.exit(1)

    log_s = np.log2(np.array(fit_sizes, dtype=float))
    log_n = np.log2(np.array([medians[s] for s in fit_sizes]))
    A = np.column_stack([np.ones(len(log_s)), log_s])
    coeffs = np.linalg.lstsq(A, log_n, rcond=None)[0]
    C = 2**coeffs[0]
    alpha = coeffs[1]

    print(f"Power-law fit (>= 64B): ns = {C:.4f} * size^{alpha:.4f}")

    # Build 32-entry table
    table_ns = []
    for k in range(31):
        s = 1 << k
        if k < 6:
            table_ns.append(exact_ns[k])
        else:
            table_ns.append(C * s**alpha)
    table_ns.append(table_ns[30])  # entry [31] = capped at 1GB

    # Enforce monotonic
    for i in range(1, len(table_ns)):
        if table_ns[i] < table_ns[i-1]:
            table_ns[i] = table_ns[i-1]

    return table_ns, medians, C, alpha

def print_go_table(table_ns, cpu_base_ns):
    """Print Go source for allocGasTable."""
    print(f"\n// Calibrated with cpuBaseNs = {cpu_base_ns}")
    print("var allocGasTable = [32]int64{")
    for k in range(32):
        s = 1 << min(k, 30)
        mg = int(round(table_ns[k] / cpu_base_ns * 1000))
        if s < 1024:
            label = f"{s}B"
        elif s < 1048576:
            label = f"{s/1024:.0f}KB"
        elif s < 1073741824:
            label = f"{s/1048576:.0f}MB"
        else:
            label = f"{s/1073741824:.0f}GB"
        src = "exact" if k < 6 else "power-law"
        capped = " (capped)" if k == 31 else ""
        print(f"\t{mg:>12}, // 2^{k:<2} = {label:>8}   ({table_ns[k]:.0f}ns, {src}{capped})")
    print("}")

def print_accuracy(table_ns, medians):
    """Print accuracy report comparing model to actual benchmarks."""
    sizes = sorted(medians.keys())

    def model_ns(size):
        if size <= 1:
            return table_ns[0]
        k = int(np.log2(size))
        if k >= 31:
            return table_ns[31]
        lo = table_ns[k]
        hi = table_ns[k+1]
        frac = size - (1 << k)
        span = 1 << k
        return lo + (hi - lo) * frac / span

    print(f"\n{'Size':>12} {'Actual ns':>10} {'Model ns':>10} {'Ratio':>8}")
    print("-" * 44)
    worst_over, worst_under = 1.0, 1.0
    for s in sizes:
        m = model_ns(s)
        r = medians[s] / m
        worst_over = max(worst_over, r)
        worst_under = min(worst_under, r)
        label = f"{s}B" if s < 1024 else f"{s/1024:.0f}KB" if s < 1048576 else f"{s/1048576:.0f}MB" if s < 1073741824 else f"{s/1073741824:.0f}GB"
        flag = " <<<" if r > 1.3 or r < 0.7 else ""
        print(f"{label:>12} {medians[s]:>10.1f} {m:>10.1f} {r:>8.2f}{flag}")

    print(f"\nWorst overcharge: {worst_under:.2f}x (model > actual)")
    print(f"Worst undercharge: {worst_over:.2f}x (actual > model)")

def plot(table_ns, medians, C, alpha, output_path):
    """Generate comparison plot."""
    try:
        import matplotlib.pyplot as plt
    except ImportError:
        print("matplotlib not installed, skipping plot", file=sys.stderr)
        return

    sizes = sorted(medians.keys())

    def model_ns(size):
        if size <= 1:
            return table_ns[0]
        k = int(np.log2(size))
        if k >= 31:
            return table_ns[31]
        lo = table_ns[k]
        hi = table_ns[k+1]
        frac = size - (1 << k)
        span = 1 << k
        return lo + (hi - lo) * frac / span

    model = [model_ns(s) for s in sizes]
    ratios = [medians[s] / model_ns(s) for s in sizes]

    fig, (ax1, ax2) = plt.subplots(2, 1, figsize=(12, 10))

    ax1.loglog(sizes, [medians[s] for s in sizes], 'bo-', markersize=5, label='Actual (median)', linewidth=2)
    ax1.loglog(sizes, model, 'r^--', markersize=4,
               label=f'Table(6) + power law ({C:.2f} × size^{alpha:.3f})', linewidth=1.5)
    pw_x = np.logspace(np.log10(64), np.log10(2e9), 200)
    ax1.loglog(pw_x, C * pw_x**alpha, 'k:', alpha=0.3, label='Pure power law')
    ax1.axvline(x=32, color='gray', linestyle=':', alpha=0.5, label='Table → power-law (32B)')
    ax1.set_xlabel('Allocation size (bytes)')
    ax1.set_ylabel('Time (ns)')
    ax1.set_title(f'Allocation gas model: 6-point table + power law (α={alpha:.3f})')
    ax1.legend(fontsize=10)
    ax1.grid(True, which='both', alpha=0.3)

    ax2.semilogx(sizes, ratios, 'go-', markersize=5)
    ax2.axhline(y=1.0, color='k', linestyle='--', alpha=0.5)
    ax2.axhline(y=1.5, color='orange', linestyle=':', alpha=0.3, label='±50%')
    ax2.axhline(y=0.5, color='orange', linestyle=':', alpha=0.3)
    ax2.set_xlabel('Allocation size (bytes)')
    ax2.set_ylabel('Ratio (actual / model)')
    ax2.set_title('Model accuracy (1.0 = perfect)')
    ax2.set_ylim(0.3, 2.0)
    ax2.legend()
    ax2.grid(True, which='both', alpha=0.3)

    for i, s in enumerate(sizes):
        if abs(ratios[i] - 1.0) > 0.25:
            label = f"{s}B" if s < 1024 else f"{s/1024:.0f}KB" if s < 1048576 else f"{s/1048576:.0f}MB"
            ax2.annotate(f'{label}\n{ratios[i]:.2f}', (s, ratios[i]),
                        textcoords="offset points", xytext=(0, 10), fontsize=8, ha='center')

    plt.tight_layout()
    plt.savefig(output_path, dpi=150)
    print(f"\nPlot saved to {output_path}")

def main():
    parser = argparse.ArgumentParser(description="Generate allocGasTable from Go allocation benchmarks")
    parser.add_argument("bench_file", help="Path to `go test -bench` output file")
    parser.add_argument("--cpu-base-ns", type=float, default=10.0,
                       help="Nanoseconds per gas unit (from benchops). Default: 10.0")
    parser.add_argument("--plot", default="alloc_gas_model.png",
                       help="Output plot path. Default: alloc_gas_model.png")
    parser.add_argument("--no-plot", action="store_true", help="Skip plot generation")
    args = parser.parse_args()

    data = parse_benchmarks(args.bench_file)
    if not data:
        print(f"ERROR: no BenchmarkAlloc entries found in {args.bench_file}", file=sys.stderr)
        sys.exit(1)

    print(f"Parsed {sum(len(v) for v in data.values())} runs across {len(data)} sizes")

    table_ns, medians, C, alpha = build_table(data)
    print_go_table(table_ns, args.cpu_base_ns)
    print_accuracy(table_ns, medians)

    if not args.no_plot:
        plot(table_ns, medians, C, alpha, args.plot)

if __name__ == "__main__":
    main()
