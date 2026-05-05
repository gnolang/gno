#!/usr/bin/env python3
"""
Generate native function gas table for GnoVM stdlibs from Go benchmarks.

Mirrors gen_analysis.py (opcode handler gas) and gen_alloc_table.py
(allocation gas). Reads `go test -bench=BenchmarkNative` output, fits a
formula per native (flat or base + slope*N), and prints:

  - native_gas_formulas.md  — markdown table of fits
  - native_gas_table.go.txt — Go-pasteable nativeGasTable block
  - native_gas_fits.png      — multi-panel log-log plot of every linear fit

Convention: 1 gas = 1 ns on reference hardware. Slope is emitted as ns
per 1024 units of N (precedent: incrCPUBigInt in
gnovm/pkg/gnolang/machine.go uses `slopePerKb / 1024`).

Usage:
    cd gnovm/cmd/calibrate
    go test -bench=BenchmarkNative -count=3 -benchtime=200ms -timeout=20m . \
        > native_bench_output.txt
    python3 gen_native_table.py native_bench_output.txt
"""
import argparse
import re
import sys
from collections import defaultdict

import numpy as np


# Per-native spec.
#   pkg, fn:               registry key (matches stdlibs/native_gas.go)
#   slope_idx:             param index for slope (None for flat,
#                          -1 for kinds that don't index a param)
#   slope_kind:            SizeKind name
#   bench_re:              regex; for variable-cost captures (size, ns),
#                          for flat captures (ns,)
NATIVE_SPECS = [
    # ---- pure CPU, slope on input bytes ----
    ("crypto/sha256", "sum256", 0, "LenBytes",
     r"BenchmarkNative_SHA256_Sum256_(\d+)-\d+\s+\d+\s+([\d.]+)\s+ns/op"),
    ("crypto/ed25519", "verify", 1, "LenBytes",
     r"BenchmarkNative_Ed25519_Verify_(\d+)-\d+\s+\d+\s+([\d.]+)\s+ns/op"),
    ("chain", "packageAddress", 0, "LenString",
     r"BenchmarkNative_Chain_PackageAddress_(\d+)-\d+\s+\d+\s+([\d.]+)\s+ns/op"),
    ("chain", "deriveStorageDepositAddr", 0, "LenString",
     r"BenchmarkNative_Chain_DeriveStorageDepositAddr_(\d+)-\d+\s+\d+\s+([\d.]+)\s+ns/op"),

    # ---- flat ----
    ("chain", "pubKeyAddress", None, "Flat",
     r"BenchmarkNative_Chain_PubKeyAddress-\d+\s+\d+\s+([\d.]+)\s+ns/op"),
    ("time", "loadFromEmbeddedTZData", None, "Flat",
     r"BenchmarkNative_Time_LoadTZData-\d+\s+\d+\s+([\d.]+)\s+ns/op"),
    ("math", "Float32bits", None, "Flat",
     r"BenchmarkNative_Math_Float32bits-\d+\s+\d+\s+([\d.]+)\s+ns/op"),
    ("math", "Float32frombits", None, "Flat",
     r"BenchmarkNative_Math_Float32frombits-\d+\s+\d+\s+([\d.]+)\s+ns/op"),
    ("math", "Float64bits", None, "Flat",
     r"BenchmarkNative_Math_Float64bits-\d+\s+\d+\s+([\d.]+)\s+ns/op"),
    ("math", "Float64frombits", None, "Flat",
     r"BenchmarkNative_Math_Float64frombits-\d+\s+\d+\s+([\d.]+)\s+ns/op"),

    # ---- chain/banker ----
    ("chain/banker", "bankerSendCoins", 3, "LenSlice",
     r"BenchmarkNative_Banker_SendCoins_(\d+)-\d+\s+\d+\s+([\d.]+)\s+ns/op"),
    ("chain/banker", "bankerGetCoins", 1, "ReturnLen",
     r"BenchmarkNative_Banker_GetCoins_(\d+)-\d+\s+\d+\s+([\d.]+)\s+ns/op"),
    ("chain/banker", "bankerTotalCoin", None, "Flat",
     r"BenchmarkNative_Banker_TotalCoin-\d+\s+\d+\s+([\d.]+)\s+ns/op"),
    ("chain/banker", "bankerIssueCoin", None, "Flat",
     r"BenchmarkNative_Banker_IssueCoin-\d+\s+\d+\s+([\d.]+)\s+ns/op"),
    ("chain/banker", "bankerRemoveCoin", None, "Flat",
     r"BenchmarkNative_Banker_RemoveCoin-\d+\s+\d+\s+([\d.]+)\s+ns/op"),
    ("chain/banker", "originSend", None, "Flat",
     r"BenchmarkNative_Banker_OriginSend-\d+\s+\d+\s+([\d.]+)\s+ns/op"),
    ("chain/banker", "assertCallerIsRealm", None, "Flat",
     r"BenchmarkNative_Banker_AssertCallerIsRealm-\d+\s+\d+\s+([\d.]+)\s+ns/op"),

    # ---- chain.emit ----
    ("chain", "emit", 1, "LenSlice",
     r"BenchmarkNative_Chain_Emit_(\d+)-\d+\s+\d+\s+([\d.]+)\s+ns/op"),

    # ---- chain/params (sets only; payload-bytes slope where applicable) ----
    ("chain/params", "SetBytes", 1, "LenBytes",
     r"BenchmarkNative_Params_SetBytes_(\d+)-\d+\s+\d+\s+([\d.]+)\s+ns/op"),
    ("chain/params", "SetString", 1, "LenString",
     r"BenchmarkNative_Params_SetString_(\d+)-\d+\s+\d+\s+([\d.]+)\s+ns/op"),
    ("chain/params", "SetStrings", 1, "LenSlice",
     r"BenchmarkNative_Params_SetStrings_(\d+)-\d+\s+\d+\s+([\d.]+)\s+ns/op"),
    ("chain/params", "SetBool", None, "Flat",
     r"BenchmarkNative_Params_SetBool-\d+\s+\d+\s+([\d.]+)\s+ns/op"),
    ("chain/params", "SetInt64", None, "Flat",
     r"BenchmarkNative_Params_SetInt64-\d+\s+\d+\s+([\d.]+)\s+ns/op"),
    ("chain/params", "SetUint64", None, "Flat",
     r"BenchmarkNative_Params_SetUint64-\d+\s+\d+\s+([\d.]+)\s+ns/op"),

    # ---- sys/params ----
    ("sys/params", "setSysParamBytes", 4, "LenBytes",
     r"BenchmarkNative_SysParams_SetBytes_(\d+)-\d+\s+\d+\s+([\d.]+)\s+ns/op"),
    ("sys/params", "getSysParamBytes", 0, "ReturnLen",
     r"BenchmarkNative_SysParams_GetBytes_(\d+)-\d+\s+\d+\s+([\d.]+)\s+ns/op"),
    ("sys/params", "setSysParamString", 4, "LenString",
     r"BenchmarkNative_SysParams_SetString_(\d+)-\d+\s+\d+\s+([\d.]+)\s+ns/op"),
    ("sys/params", "setSysParamStrings", 4, "LenSlice",
     r"BenchmarkNative_SysParams_SetStrings_(\d+)-\d+\s+\d+\s+([\d.]+)\s+ns/op"),
    ("sys/params", "updateSysParamStrings", 4, "LenSlice",
     r"BenchmarkNative_SysParams_UpdateStrings_(\d+)-\d+\s+\d+\s+([\d.]+)\s+ns/op"),
    ("sys/params", "setSysParamBool", None, "Flat",
     r"BenchmarkNative_SysParams_SetBool-\d+\s+\d+\s+([\d.]+)\s+ns/op"),
    ("sys/params", "setSysParamInt64", None, "Flat",
     r"BenchmarkNative_SysParams_SetInt64-\d+\s+\d+\s+([\d.]+)\s+ns/op"),
    ("sys/params", "setSysParamUint64", None, "Flat",
     r"BenchmarkNative_SysParams_SetUint64-\d+\s+\d+\s+([\d.]+)\s+ns/op"),
    ("sys/params", "getSysParamBool", None, "Flat",
     r"BenchmarkNative_SysParams_GetBool-\d+\s+\d+\s+([\d.]+)\s+ns/op"),
    ("sys/params", "getSysParamInt64", None, "Flat",
     r"BenchmarkNative_SysParams_GetInt64-\d+\s+\d+\s+([\d.]+)\s+ns/op"),
    ("sys/params", "getSysParamUint64", None, "Flat",
     r"BenchmarkNative_SysParams_GetUint64-\d+\s+\d+\s+([\d.]+)\s+ns/op"),
    ("sys/params", "getSysParamString", 0, "ReturnLen",
     r"BenchmarkNative_SysParams_GetString_(\d+)-\d+\s+\d+\s+([\d.]+)\s+ns/op"),
    ("sys/params", "getSysParamStrings", 0, "ReturnLen",
     r"BenchmarkNative_SysParams_GetStrings_(\d+)-\d+\s+\d+\s+([\d.]+)\s+ns/op"),
    ("chain/params", "UpdateParamStrings", 1, "LenSlice",
     r"BenchmarkNative_Params_UpdateStrings_(\d+)-\d+\s+\d+\s+([\d.]+)\s+ns/op"),

    # ---- chain/runtime ----
    ("chain/runtime", "ChainID", None, "Flat",
     r"BenchmarkNative_Runtime_ChainID-\d+\s+\d+\s+([\d.]+)\s+ns/op"),
    ("chain/runtime", "ChainDomain", None, "Flat",
     r"BenchmarkNative_Runtime_ChainDomain-\d+\s+\d+\s+([\d.]+)\s+ns/op"),
    ("chain/runtime", "ChainHeight", None, "Flat",
     r"BenchmarkNative_Runtime_ChainHeight-\d+\s+\d+\s+([\d.]+)\s+ns/op"),
    ("chain/runtime", "originCaller", None, "Flat",
     r"BenchmarkNative_Runtime_OriginCaller-\d+\s+\d+\s+([\d.]+)\s+ns/op"),
    ("chain/runtime", "getSessionInfo", None, "Flat",
     r"BenchmarkNative_Runtime_GetSessionInfo-\d+\s+\d+\s+([\d.]+)\s+ns/op"),
    ("chain/runtime", "AssertOriginCall", None, "Flat",
     r"BenchmarkNative_Runtime_AssertOriginCall-\d+\s+\d+\s+([\d.]+)\s+ns/op"),
    ("chain/runtime", "getRealm", -1, "NumFrames",
     r"BenchmarkNative_Runtime_GetRealm_(\d+)-\d+\s+\d+\s+([\d.]+)\s+ns/op"),

    # ---- time ----
    ("time", "now", None, "Flat",
     r"BenchmarkNative_Time_Now-\d+\s+\d+\s+([\d.]+)\s+ns/op"),
]


def parse_bench(path):
    text = open(path).read()
    var_data = defaultdict(lambda: defaultdict(list))
    flat_data = defaultdict(list)
    for pkg, fn, slope_idx, kind, regex in NATIVE_SPECS:
        for line in text.splitlines():
            m = re.search(regex, line)
            if not m:
                continue
            if kind == "Flat":
                flat_data[(pkg, fn)].append(float(m.group(1)))
            else:
                size = int(m.group(1))
                ns = float(m.group(2))
                var_data[(pkg, fn)][size].append(ns)
    return var_data, flat_data


def fit_linear(sizes_ns):
    """Weighted LS (1/y) so small N stays relevant; floor base to ns(min)."""
    sizes = np.array(sorted(sizes_ns.keys()), dtype=float)
    ns = np.array([np.median(sizes_ns[s]) for s in sizes], dtype=float)
    if len(sizes) < 2:
        return float(ns[0]), 0.0, 1.0
    w = 1.0 / np.maximum(ns, 1e-9)
    A = np.column_stack([np.ones(len(sizes)), sizes])
    AW = (A.T * w).T
    yW = ns * w
    coeffs, *_ = np.linalg.lstsq(AW, yW, rcond=None)
    base = max(float(coeffs[0]), float(ns[0]))
    slope = max(float(coeffs[1]), 0.0)
    pred = base + slope * sizes
    ss_res = float(np.sum((ns - pred) ** 2))
    ss_tot = float(np.sum((ns - ns.mean()) ** 2))
    r2 = 1.0 - ss_res / ss_tot if ss_tot > 0 else 1.0
    return base, slope, r2


def fit_flat(values):
    return float(np.median(values))


def n_desc(kind, slope_idx):
    return {
        "LenBytes": f"len(p{slope_idx}) bytes",
        "LenString": f"len(p{slope_idx}) string",
        "LenSlice": f"len(p{slope_idx}) slice",
        "NumFrames": "m.NumFrames()",
        "ReturnLen": f"len(return[{slope_idx}])",
        "SumLenStrings": f"sum_len(p{slope_idx})",
    }[kind]


def emit_markdown(rows, out):
    out.write("# Native Function Gas Formulas\n\n")
    out.write("Generated by `gen_native_table.py`. 1 gas = 1 ns on reference hardware.\n")
    out.write("Slope is ns/N; runtime stores it as `Slope/1024` and computes `base + slope*N/1024`.\n\n")
    out.write("| Native | Shape | Base (ns) | Slope (ns/N) | N | R² |\n")
    out.write("|---|---|---:|---:|---|---:|\n")
    for r in rows:
        if r["shape"] == "flat":
            out.write(f"| `{r['pkg']}.{r['fn']}` | flat | {r['base']:.1f} | — | — | — |\n")
        else:
            out.write(f"| `{r['pkg']}.{r['fn']}` | base+slope·N | {r['base']:.1f} | {r['slope']:.4f} | {n_desc(r['kind'], r['slope_idx'])} | {r['r2']:.3f} |\n")


def emit_go_table(rows, out):
    out.write("// Code generated by gen_native_table.py from native_bench_output.txt.\n")
    out.write("// 1 gas = 1 ns on reference hardware (Intel Xeon Platinum 8168).\n")
    out.write("// Slope is ns per 1024 units of N; runtime computes base + slope*N/1024.\n")
    out.write("// See gnovm/cmd/calibrate/native_gas_formulas.md for derivation.\n")
    out.write("var calibratedNativeGas = []NativeGasEntry{\n")
    for r in rows:
        if r["shape"] == "flat":
            out.write(
                f'\t{{Pkg: "{r["pkg"]}", Fn: "{r["fn"]}", '
                f'Base: {int(round(r["base"]))}, '
                f'Slope: 0, SlopeIdx: -1, SlopeKind: SizeFlat}},'
                f' // flat, median {r["base"]:.1f}ns\n'
            )
        else:
            slope_per_1024 = int(round(r["slope"] * 1024))
            out.write(
                f'\t{{Pkg: "{r["pkg"]}", Fn: "{r["fn"]}", '
                f'Base: {int(round(r["base"]))}, '
                f'Slope: {slope_per_1024}, '
                f'SlopeIdx: {r["slope_idx"]}, '
                f'SlopeKind: Size{r["kind"]}}},'
                f' // fit base={r["base"]:.1f}ns slope={r["slope"]:.4f}ns/N (={slope_per_1024}/1024) R²={r["r2"]:.3f}\n'
            )
    out.write("}\n")


def _param_label(r):
    """Human-readable parameter description shown on each panel's x-axis.

    Linear fits get the size source (e.g. "N = len(p1) slice") so the
    reader can map the slope back to what gas charge it produces.
    Flat fits get an explicit "(flat)" annotation."""
    if r["shape"] == "flat":
        return "(flat — no size parameter)"
    kind = r["kind"]
    if kind == "NumFrames":
        return "N = m.NumFrames()"
    if kind == "ReturnLen":
        return f"N = len(return[{r['slope_idx']}])"
    if kind == "SumLenStrings":
        return f"N = sum_len(p{r['slope_idx']})"
    return f"N = len(p{r['slope_idx']}) — {kind}"


def plot_fits(var_data, flat_data, rows, out_path):
    """Render every calibrated native — both linear and flat fits.

    - Linear panels: median data points (blue) + fit line (red dashed).
    - Flat panels: median annotated as a single horizontal reference
      with the value on the title; no x-axis sweep to plot.
    Each panel's x-axis label states the parameter being measured (or
    "(flat)" for natives with no size dependence) so the reader can
    map every slope back to the runtime gas charge."""
    try:
        import matplotlib.pyplot as plt
    except ImportError:
        print("matplotlib not installed, skipping plot", file=sys.stderr)
        return
    if not rows:
        return
    n = len(rows)
    cols = 3
    rows_n = (n + cols - 1) // cols
    fig, axes = plt.subplots(rows_n, cols,
                             figsize=(5 * cols, 3.2 * rows_n), squeeze=False)
    for i, r in enumerate(rows):
        ax = axes[i // cols][i % cols]
        param = _param_label(r)

        if r["shape"] == "flat":
            base = r["base"]
            ax.axhline(base, color="green", linestyle="-", linewidth=2,
                       label=f"flat = {base:.1f} ns")
            ax.set_xlim(0, 1)
            ax.set_xticks([])
            ax.set_ylim(0, max(base * 2.0, 10))
            ax.set_title(f"{r['pkg']}.{r['fn']}\n→ {base:.1f} ns flat",
                         fontsize=9)
        else:
            d = var_data[(r["pkg"], r["fn"])]
            sizes = np.array(sorted(d.keys()), dtype=float)
            med = np.array([np.median(d[s]) for s in sizes])
            ax.plot(sizes, med, "bo-", markersize=5, label="median ns/op")
            if sizes.min() <= 0:
                xs = np.linspace(0, sizes.max(), 200)
            else:
                xs = np.geomspace(sizes.min(), sizes.max(), 200)
            ys = r["base"] + r["slope"] * xs
            ax.plot(xs, ys, "r--",
                    label=f"fit: {r['base']:.0f}+{r['slope']:.3f}·N  R²={r['r2']:.3f}")
            if sizes.min() > 0:
                ax.set_xscale("log", base=10)
                ax.set_yscale("log")
            ax.set_title(f"{r['pkg']}.{r['fn']}", fontsize=9)

        ax.set_xlabel(param, fontsize=8)
        ax.set_ylabel("ns/op", fontsize=8)
        ax.legend(fontsize=7, loc="best")
        ax.grid(True, which="both", alpha=0.3)

    for j in range(n, rows_n * cols):
        axes[j // cols][j % cols].axis("off")
    plt.tight_layout()
    plt.savefig(out_path, dpi=140)
    print(f"\nPlot saved to {out_path}", file=sys.stderr)


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("bench_file")
    ap.add_argument("--md-out", default="native_gas_formulas.md")
    ap.add_argument("--go-out", default="native_gas_table.go.txt")
    ap.add_argument("--plot", default="native_gas_fits.png")
    ap.add_argument("--no-plot", action="store_true")
    args = ap.parse_args()

    var_data, flat_data = parse_bench(args.bench_file)
    rows = []
    for pkg, fn, slope_idx, kind, _ in NATIVE_SPECS:
        if kind == "Flat":
            vs = flat_data.get((pkg, fn))
            if not vs:
                print(f"WARN: no data for flat {pkg}.{fn}", file=sys.stderr)
                continue
            rows.append({"pkg": pkg, "fn": fn, "shape": "flat",
                         "base": fit_flat(vs)})
        else:
            d = var_data.get((pkg, fn))
            if not d or len(d) < 2:
                print(f"WARN: not enough size points for {pkg}.{fn}", file=sys.stderr)
                continue
            base, slope, r2 = fit_linear(d)
            # Demote to flat if per-1024 slope rounds to 0 — those are
            # natives whose X_ wrapper does no per-N work (e.g. params
            # writes; per-byte cost lives in the KVStore layer).
            if int(round(slope * 1024)) == 0:
                rows.append({"pkg": pkg, "fn": fn, "shape": "flat",
                             "base": base})
                print(f"NOTE: {pkg}.{fn} demoted to flat (slope rounded to 0)", file=sys.stderr)
            else:
                rows.append({"pkg": pkg, "fn": fn, "shape": "linear",
                             "base": base, "slope": slope, "r2": r2,
                             "slope_idx": slope_idx, "kind": kind})

    with open(args.md_out, "w") as f:
        emit_markdown(rows, f)
    print(f"Wrote {args.md_out}", file=sys.stderr)
    with open(args.go_out, "w") as f:
        emit_go_table(rows, f)
    print(f"Wrote {args.go_out}", file=sys.stderr)
    print()
    emit_go_table(rows, sys.stdout)
    if not args.no_plot:
        plot_fits(var_data, flat_data, rows, args.plot)


if __name__ == "__main__":
    main()
