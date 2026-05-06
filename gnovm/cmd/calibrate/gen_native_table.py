#!/usr/bin/env python3
"""
Generate native function gas table for GnoVM stdlibs from Go benchmarks.

Mirrors gen_analysis.py (opcode handler gas) and gen_alloc_table.py
(allocation gas). Reads `go test -bench=BenchmarkNative` output, fits a
formula per native (flat, base + slope*N, or base + α·count + β·total_bytes
for slice-of-string natives), and prints:

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

    # ---- chain/banker (denom strings small; per-coin slope only) ----
    ("chain/banker", "bankerSendCoins", 3, "LenSlice",
     r"BenchmarkNative_Banker_SendCoins_(\d+)-\d+\s+\d+\s+([\d.]+)\s+ns/op"),
    # SizeReturnLen → post-call charge. slope_idx is stack offset from
    # top (1-based): bankerGetCoins → (denoms, amounts), denoms is the
    # first-declared return so it lands at stack offset 2 (amounts at 1).
    ("chain/banker", "bankerGetCoins", 2, "ReturnLen",
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

    # ---- chain/params (sets only; payload-bytes slope where applicable) ----
    ("chain/params", "SetBytes", 1, "LenBytes",
     r"BenchmarkNative_Params_SetBytes_(\d+)-\d+\s+\d+\s+([\d.]+)\s+ns/op"),
    ("chain/params", "SetString", 1, "LenString",
     r"BenchmarkNative_Params_SetString_(\d+)-\d+\s+\d+\s+([\d.]+)\s+ns/op"),
    ("chain/params", "SetBool", None, "Flat",
     r"BenchmarkNative_Params_SetBool-\d+\s+\d+\s+([\d.]+)\s+ns/op"),
    ("chain/params", "SetInt64", None, "Flat",
     r"BenchmarkNative_Params_SetInt64-\d+\s+\d+\s+([\d.]+)\s+ns/op"),
    ("chain/params", "SetUint64", None, "Flat",
     r"BenchmarkNative_Params_SetUint64-\d+\s+\d+\s+([\d.]+)\s+ns/op"),

    # ---- sys/params ----
    # Setters share the (module, submodule, name, val[, add]) shape — the
    # variable-cost slope param sits at index 3 (NOT 4: name comes BEFORE
    # val). Bench harness still uses 4-5 block slots; the SlopeIdx below is
    # what production chargeNativeGas indexes into at runtime.
    ("sys/params", "setSysParamBytes", 3, "LenBytes",
     r"BenchmarkNative_SysParams_SetBytes_(\d+)-\d+\s+\d+\s+([\d.]+)\s+ns/op"),
    # getSysParamBytes → (value, found). value is first-declared, stack offset 2.
    ("sys/params", "getSysParamBytes", 2, "ReturnLen",
     r"BenchmarkNative_SysParams_GetBytes_(\d+)-\d+\s+\d+\s+([\d.]+)\s+ns/op"),
    ("sys/params", "setSysParamString", 3, "LenString",
     r"BenchmarkNative_SysParams_SetString_(\d+)-\d+\s+\d+\s+([\d.]+)\s+ns/op"),
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
    ("sys/params", "getSysParamString", 2, "ReturnLen",
     r"BenchmarkNative_SysParams_GetString_(\d+)-\d+\s+\d+\s+([\d.]+)\s+ns/op"),

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
    ("chain/runtime", "getRealm", -1, "NumCallFrames",
     r"BenchmarkNative_Runtime_GetRealm_(\d+)-\d+\s+\d+\s+([\d.]+)\s+ns/op"),

    # ---- time ----
    ("time", "now", None, "Flat",
     r"BenchmarkNative_Time_Now-\d+\s+\d+\s+([\d.]+)\s+ns/op"),
]


# 2D specs: natives whose cost depends on BOTH element count AND total
# inner bytes (e.g. chain.emit, *.SetStrings). Bench names are
# "_<count>_<perElemBytes>-<gomaxprocs>" and the fitter regresses
#   cost = base + α·count + β·(count*perElemBytes)
# producing a NativeGasInfo with two additive slopes (Slope on count,
# Slope2 on SliceTotalBytes). Post-call entries (post=True) emit into
# the PostSlope/PostSlope2 fields and skip the pre-call slopes.
#
# Format: (pkg, fn, idx, count_kind, regex, post_call)
NATIVE_SPECS_2D = [
    ("chain", "emit", 1, "LenSlice",
     r"BenchmarkNative_Chain_Emit_(\d+)_(\d+)-\d+\s+\d+\s+([\d.]+)\s+ns/op", False),
    ("chain/params", "SetStrings", 1, "LenSlice",
     r"BenchmarkNative_Params_SetStrings_(\d+)_(\d+)-\d+\s+\d+\s+([\d.]+)\s+ns/op", False),
    ("chain/params", "UpdateParamStrings", 1, "LenSlice",
     r"BenchmarkNative_Params_UpdateStrings_(\d+)_(\d+)-\d+\s+\d+\s+([\d.]+)\s+ns/op", False),
    ("sys/params", "setSysParamStrings", 3, "LenSlice",
     r"BenchmarkNative_SysParams_SetStrings_(\d+)_(\d+)-\d+\s+\d+\s+([\d.]+)\s+ns/op", False),
    ("sys/params", "updateSysParamStrings", 3, "LenSlice",
     r"BenchmarkNative_SysParams_UpdateStrings_(\d+)_(\d+)-\d+\s+\d+\s+([\d.]+)\s+ns/op", False),
    ("sys/params", "getSysParamStrings", 2, "ReturnLen",
     r"BenchmarkNative_SysParams_GetStrings_(\d+)_(\d+)-\d+\s+\d+\s+([\d.]+)\s+ns/op", True),
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


def parse_bench_2d(path):
    """Parse 2-D bench rows: (count, perElem) → list of ns observations."""
    text = open(path).read()
    data = defaultdict(lambda: defaultdict(list))
    for pkg, fn, _, _, regex, _ in NATIVE_SPECS_2D:
        for line in text.splitlines():
            m = re.search(regex, line)
            if not m:
                continue
            count = int(m.group(1))
            per_elem = int(m.group(2))
            ns = float(m.group(3))
            data[(pkg, fn)][(count, per_elem)].append(ns)
    return data


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


def fit_2d(grid):
    """Multivariate LS for cost = base + α·count + β·total_bytes.

    grid: dict (count, per_elem) → list of ns observations.
    Returns (base, alpha_count, beta_bytes, r2).
    Weighted by 1/y to keep small/cheap data points from dominating.
    Coefficients are floored at zero (gas can't be negative)."""
    pts = []
    for (c, p), nss in grid.items():
        pts.append((c, c * p, float(np.median(nss))))
    if len(pts) < 3:
        # Need at least 3 to disambiguate base + 2 slopes.
        return None
    pts.sort()
    arr = np.array(pts, dtype=float)
    counts = arr[:, 0]
    bytes_ = arr[:, 1]
    ns = arr[:, 2]
    w = 1.0 / np.maximum(ns, 1e-9)
    A = np.column_stack([np.ones(len(arr)), counts, bytes_])
    AW = (A.T * w).T
    yW = ns * w
    coeffs, *_ = np.linalg.lstsq(AW, yW, rcond=None)
    base = max(float(coeffs[0]), float(ns.min()) * 0.5)
    alpha = max(float(coeffs[1]), 0.0)
    beta = max(float(coeffs[2]), 0.0)
    pred = base + alpha * counts + beta * bytes_
    ss_res = float(np.sum((ns - pred) ** 2))
    ss_tot = float(np.sum((ns - ns.mean()) ** 2))
    r2 = 1.0 - ss_res / ss_tot if ss_tot > 0 else 1.0
    return base, alpha, beta, r2


def fit_flat(values):
    return float(np.median(values))


def n_desc(kind, slope_idx):
    return {
        "LenBytes":         f"len(p{slope_idx}) bytes",
        "LenString":        f"len(p{slope_idx}) string",
        "LenSlice":         f"len(p{slope_idx}) slice",
        "NumCallFrames":    "m.NumCallFrames()",
        "ReturnLen":        f"len(return[{slope_idx}])",
        "SliceTotalBytes":  f"sum_inner_len(p{slope_idx})",
    }[kind]


def emit_markdown(rows, out):
    out.write("# Native Function Gas Formulas\n\n")
    out.write("Generated by `gen_native_table.py`. 1 gas = 1 ns on reference hardware.\n")
    out.write("Slope is ns/N; runtime stores it as `Slope/1024` and computes `base + slope*N/1024`.\n\n")
    out.write("| Native | Shape | Base (ns) | α (ns/elem) | β (ns/byte) | N | R² |\n")
    out.write("|---|---|---:|---:|---:|---|---:|\n")
    for r in rows:
        if r["shape"] == "flat":
            out.write(f"| `{r['pkg']}.{r['fn']}` | flat | {r['base']:.1f} | — | — | — | — |\n")
        elif r["shape"] == "linear":
            out.write(
                f"| `{r['pkg']}.{r['fn']}` | base+α·N | {r['base']:.1f} | "
                f"{r['slope']:.4f} | — | {n_desc(r['kind'], r['slope_idx'])} | {r['r2']:.3f} |\n"
            )
        elif r["shape"] == "2d":
            out.write(
                f"| `{r['pkg']}.{r['fn']}` | base+α·count+β·bytes | "
                f"{r['base']:.1f} | {r['alpha']:.4f} | {r['beta']:.4f} | "
                f"{n_desc(r['count_kind'], r['idx'])} + sum_inner_len | {r['r2']:.3f} |\n"
            )


def emit_go_table(rows, out):
    """Emit nativeGasEntry literals matching gno.NativeGasInfo's shape.

    SizeReturnLen (1-D) emits as post-call charges (Base flat + PostBase/
    PostSlope/PostSlopeIdx/PostSlopeKind). 2-D fits emit Slope+Slope2
    (or PostSlope+PostSlope2 when post-call) in one entry."""
    out.write("// Code generated by gen_native_table.py from native_bench_output.txt.\n")
    out.write("// 1 gas = 1 ns on reference hardware (Intel Xeon Platinum 8168).\n")
    out.write("// Slope is ns per 1024 units of N; runtime computes base + slope*N/1024.\n")
    out.write("// See gnovm/cmd/calibrate/native_gas_formulas.md for derivation.\n")
    out.write("var calibratedNativeGas = []nativeGasEntry{\n")
    for r in rows:
        if r["shape"] == "flat":
            out.write(
                f'\t{{Pkg: "{r["pkg"]}", Fn: "{r["fn"]}", '
                f'Base: {int(round(r["base"]))}, '
                f'SlopeIdx: -1, SlopeKind: SizeFlat}},'
                f' // flat, median {r["base"]:.1f}ns\n'
            )
            continue
        if r["shape"] == "2d":
            base = int(round(r["base"]))
            alpha_per_1024 = int(round(r["alpha"] * 1024))
            beta_per_1024 = int(round(r["beta"] * 1024))
            if r["post"]:
                # Post-call 2-D: pre is flat 0, post carries both slopes.
                # The pre-call base is folded into PostBase since the
                # bench measured end-to-end.
                out.write(
                    f'\t{{Pkg: "{r["pkg"]}", Fn: "{r["fn"]}", '
                    f'Base: 0, SlopeIdx: -1, SlopeKind: SizeFlat, '
                    f'PostBase: {base}, '
                    f'PostSlope: {alpha_per_1024}, '
                    f'PostSlopeIdx: {r["idx"]}, '
                    f'PostSlopeKind: Size{r["count_kind"]}, '
                    f'PostSlope2: {beta_per_1024}, '
                    f'PostSlope2Idx: {r["idx"]}, '
                    f'PostSlope2Kind: SizeSliceTotalBytes}},'
                    f' // post-2d: base={r["base"]:.1f}ns + α={r["alpha"]:.4f}ns/elem (={alpha_per_1024}/1024)'
                    f' + β={r["beta"]:.4f}ns/byte (={beta_per_1024}/1024) R²={r["r2"]:.3f}\n'
                )
            else:
                out.write(
                    f'\t{{Pkg: "{r["pkg"]}", Fn: "{r["fn"]}", '
                    f'Base: {base}, '
                    f'Slope: {alpha_per_1024}, '
                    f'SlopeIdx: {r["idx"]}, '
                    f'SlopeKind: Size{r["count_kind"]}, '
                    f'Slope2: {beta_per_1024}, '
                    f'Slope2Idx: {r["idx"]}, '
                    f'Slope2Kind: SizeSliceTotalBytes}},'
                    f' // 2d: base={r["base"]:.1f}ns + α={r["alpha"]:.4f}ns/elem (={alpha_per_1024}/1024)'
                    f' + β={r["beta"]:.4f}ns/byte (={beta_per_1024}/1024) R²={r["r2"]:.3f}\n'
                )
            continue
        # 1-D linear.
        slope_per_1024 = int(round(r["slope"] * 1024))
        if r["kind"] == "ReturnLen":
            out.write(
                f'\t{{Pkg: "{r["pkg"]}", Fn: "{r["fn"]}", '
                f'Base: {int(round(r["base"]))}, '
                f'SlopeIdx: -1, SlopeKind: SizeFlat, '
                f'PostSlope: {slope_per_1024}, '
                f'PostSlopeIdx: {r["slope_idx"]}, '
                f'PostSlopeKind: Size{r["kind"]}}},'
                f' // post-call: base={r["base"]:.1f}ns + {r["slope"]:.4f}ns/N (={slope_per_1024}/1024) R²={r["r2"]:.3f}\n'
            )
            continue
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
        return ""
    if r["shape"] == "2d":
        return f"count = len(p{r['idx']}); bytes = sum_inner_len"
    kind = r["kind"]
    if kind == "NumCallFrames":
        return "N = m.NumCallFrames()"
    if kind == "ReturnLen":
        return f"N = len(return[{r['slope_idx']}])"
    return f"N = len(p{r['slope_idx']}) — {kind}"


def plot_fits(var_data, flat_data, two_d_data, rows, out_path):
    """Render the parameterized fits — linear and 2-D.

    Flat natives (single horizontal reference; no parameter on the x-axis)
    are excluded — their plots are visually redundant with the median
    they encode, and the markdown/table already captures the value.

    - Linear panels: median data points (blue) + fit line (red dashed).
    - 2-D panels: two overlaid series — vs count (perElem=1, blue) and
      vs total bytes (count=2, orange). Each series gets the model's
      prediction (dashed) so visual deviation flags a poor fit."""
    try:
        import matplotlib.pyplot as plt
    except ImportError:
        print("matplotlib not installed, skipping plot", file=sys.stderr)
        return
    rows = [r for r in rows if r["shape"] != "flat"]
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

        if r["shape"] == "2d":
            grid = two_d_data[(r["pkg"], r["fn"])]
            # Series 1: vary count, hold per-elem at min observed (≈1).
            min_p = min(p for (_, p) in grid.keys())
            counts = sorted({c for (c, p) in grid.keys() if p == min_p})
            count_ns = [np.median(grid[(c, min_p)]) for c in counts]
            # Series 2: vary per-elem, hold count at min observed (≈2).
            min_c = min(c for (c, p) in grid.keys() if p > min_p) if any(p > min_p for (_, p) in grid.keys()) else min(c for (c, _) in grid.keys())
            elems = sorted({p for (c, p) in grid.keys() if c == min_c})
            elem_ns = [np.median(grid[(min_c, p)]) for p in elems]
            if counts and count_ns:
                ax.plot(counts, count_ns, "bo-", markersize=5,
                        label=f"vs count (perElem={min_p})")
                xs = np.array(counts, dtype=float)
                pred = r["base"] + r["alpha"] * xs + r["beta"] * xs * min_p
                ax.plot(xs, pred, "b--", alpha=0.6,
                        label=f"  α={r['alpha']:.3f}")
            if elems and elem_ns:
                # Plot bytes-axis series on the SAME axis but scaled by
                # total bytes (count*perElem) so it's comparable.
                xs2_bytes = np.array([min_c * p for p in elems], dtype=float)
                # Use the bytes-axis as a second curve; place its x as
                # total bytes so the slopes are visually disentangled.
                ax2 = ax.twiny()
                ax2.plot(xs2_bytes, elem_ns, "s-", color="orange",
                         markersize=5, label=f"vs total_bytes (count={min_c})")
                pred2 = r["base"] + r["alpha"] * min_c + r["beta"] * xs2_bytes
                ax2.plot(xs2_bytes, pred2, "--", color="orange", alpha=0.6,
                         label=f"  β={r['beta']:.3f}")
                ax2.set_xscale("log")
                ax2.set_xlabel("total bytes (count·perElem)", fontsize=7,
                               color="orange")
                ax2.tick_params(axis="x", labelcolor="orange", labelsize=7)
                ax2.legend(fontsize=7, loc="upper left")
            ax.set_xscale("log")
            ax.set_yscale("log")
            ax.set_title(
                f"{r['pkg']}.{r['fn']}\nbase={r['base']:.0f}+α·c+β·b R²={r['r2']:.3f}",
                fontsize=9)
        else:  # linear
            # 2-D entries that demoted to 1-D (β rounded to zero) live
            # in two_d_data, not var_data. Project them onto the count
            # axis (use the smallest perElem observed; that's the "vs
            # count, holding bytes near zero" slice that motivated the
            # 1-D fit).
            d = var_data.get((r["pkg"], r["fn"]))
            if d is None:
                grid = two_d_data.get((r["pkg"], r["fn"]), {})
                if not grid:
                    ax.set_title(f"{r['pkg']}.{r['fn']}\n(no data)", fontsize=9)
                    continue
                min_p = min(p for (_, p) in grid.keys())
                d = {c: grid[(c, min_p)]
                     for (c, p) in grid.keys() if p == min_p}
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

        if param:
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
    two_d_data = parse_bench_2d(args.bench_file)

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
            # Demote to flat if either the per-1024 slope rounds to 0 or
            # R² < 0.5 (the line fits worse than a constant mean).
            if int(round(slope * 1024)) == 0 or r2 < 0.5:
                rows.append({"pkg": pkg, "fn": fn, "shape": "flat",
                             "base": base})
                print(f"NOTE: {pkg}.{fn} demoted to flat (slope={slope:.6f}, R²={r2:.3f})", file=sys.stderr)
            else:
                rows.append({"pkg": pkg, "fn": fn, "shape": "linear",
                             "base": base, "slope": slope, "r2": r2,
                             "slope_idx": slope_idx, "kind": kind})

    # 2-D fits.
    for pkg, fn, idx, count_kind, _, post in NATIVE_SPECS_2D:
        grid = two_d_data.get((pkg, fn))
        if not grid:
            print(f"WARN: no 2D data for {pkg}.{fn}", file=sys.stderr)
            continue
        result = fit_2d(grid)
        if result is None:
            print(f"WARN: insufficient 2D points for {pkg}.{fn}", file=sys.stderr)
            continue
        base, alpha, beta, r2 = result
        alpha_per_1024 = int(round(alpha * 1024))
        beta_per_1024 = int(round(beta * 1024))
        # Demote noise-level slopes. Anything < 10/1024 ns/byte (≈0.01
        # ns/byte; 10µs over a 1MB payload) is below bench-to-bench
        # variance — keeping it just makes the table flip on re-runs.
        if beta_per_1024 < 10:
            beta_per_1024 = 0
        if alpha_per_1024 < 10:
            alpha_per_1024 = 0
        # Demotion ladder:
        #   both zero          → flat
        #   only β zero        → 1-D linear on count (most natives — the
        #                        per-byte cost lives in the metered KVStore,
        #                        not the dispatcher; native CPU is per-elem)
        #   only α zero        → 1-D linear on total bytes
        #   both nonzero       → 2-D
        if alpha_per_1024 == 0 and beta_per_1024 == 0:
            rows.append({"pkg": pkg, "fn": fn, "shape": "flat", "base": base})
            print(f"NOTE: {pkg}.{fn} 2D demoted to flat (α≈β≈0)", file=sys.stderr)
        elif beta_per_1024 == 0:
            # Synthesize a 1-D row matching the linear shape (kind=count).
            shape_kind = count_kind  # LenSlice / ReturnLen
            if post:
                # Post-call: 1-D variant goes through the existing
                # ReturnLen path. For LenSlice (pre-call) post=False stays.
                shape_kind = "ReturnLen"
            rows.append({"pkg": pkg, "fn": fn, "shape": "linear",
                         "base": base, "slope": alpha, "r2": r2,
                         "slope_idx": idx, "kind": shape_kind})
            print(f"NOTE: {pkg}.{fn} 2D→1D (β≈0): {alpha:.3f} ns/elem only",
                  file=sys.stderr)
        elif alpha_per_1024 == 0:
            rows.append({"pkg": pkg, "fn": fn, "shape": "linear",
                         "base": base, "slope": beta, "r2": r2,
                         "slope_idx": idx, "kind": "SliceTotalBytes"})
            print(f"NOTE: {pkg}.{fn} 2D→1D (α≈0): {beta:.4f} ns/byte only",
                  file=sys.stderr)
        else:
            rows.append({
                "pkg": pkg, "fn": fn, "shape": "2d",
                "base": base, "alpha": alpha, "beta": beta, "r2": r2,
                "idx": idx, "count_kind": count_kind, "post": post,
            })

    with open(args.md_out, "w") as f:
        emit_markdown(rows, f)
    print(f"Wrote {args.md_out}", file=sys.stderr)
    with open(args.go_out, "w") as f:
        emit_go_table(rows, f)
    print(f"Wrote {args.go_out}", file=sys.stderr)
    print()
    emit_go_table(rows, sys.stdout)
    if not args.no_plot:
        plot_fits(var_data, flat_data, two_d_data, rows, args.plot)


if __name__ == "__main__":
    main()
