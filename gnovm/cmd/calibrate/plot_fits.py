#!/usr/bin/env python3
"""Generate multi-panel plot of parameterized op benchmarks with linear fits."""

import re
import math
import matplotlib
matplotlib.use('Agg')
import matplotlib.pyplot as plt
import numpy as np

CPU_BASE_NS = 5.2

def parse_benchmarks(path):
    data = {}
    with open(path) as f:
        for line in f:
            m = re.match(
                r'^(BenchmarkOp\S+)-\d+\s+\d+\s+'
                r'([\d.]+)\s+ns/op\s+'
                r'([\d.]+)\s+alloc-gas/op\s+'
                r'([\d.]+)\s+ns/op\(pure\)', line)
            if m:
                name = m.group(1)
                ns_pure = float(m.group(4))
                alloc_gas = float(m.group(3))
                data.setdefault(name, []).append((ns_pure, alloc_gas))
    return data

def get_stats(data, name):
    if name not in data:
        return None
    runs = data[name]
    ns = sum(r[0] for r in runs) / len(runs)
    ag = sum(r[1] for r in runs) / len(runs)
    return (ns, ag)

def cpu_gas(ns_pure, alloc_gas):
    return ns_pure / CPU_BASE_NS - alloc_gas

def least_squares(points):
    n = len(points)
    if n < 2:
        return (points[0][1] if points else 0, 0, 1.0)
    sx = sum(x for x, y in points)
    sy = sum(y for x, y in points)
    sxx = sum(x*x for x, y in points)
    sxy = sum(x*y for x, y in points)
    denom = n * sxx - sx * sx
    if abs(denom) < 1e-10:
        return (sy / n, 0, 1.0)
    b = (n * sxy - sx * sy) / denom
    a = (sy - b * sx) / n
    y_mean = sy / n
    ss_tot = sum((y - y_mean)**2 for x, y in points)
    ss_res = sum((y - (a + b*x))**2 for x, y in points)
    r2 = 1 - ss_res / ss_tot if ss_tot > 0 else 1.0
    return (a, b, r2)

# Families whose cost is O(N²) because benchmarks use N1=N2=N.
# Fit against N² instead of N.
QUADRATIC_FAMILIES = {'Mul (BigInt)', 'Mul (BigDec)', 'Quo (BigDec)'}

# All parameterized families: (name, param_label, [(bench_name, N), ...])
FAMILIES = [
    ('MapLit', 'entries', [
        ('BenchmarkOpMapLit_1', 1), ('BenchmarkOpMapLit_10', 10),
        ('BenchmarkOpMapLit_100', 100), ('BenchmarkOpMapLit_1000', 1000)]),
    ('ArrayLit (int)', 'elements', [
        ('BenchmarkOpArrayLit_1', 1), ('BenchmarkOpArrayLit_10', 10),
        ('BenchmarkOpArrayLit_100', 100), ('BenchmarkOpArrayLit_1000', 1000)]),
    ('ArrayLit (uint8)', 'elements', [
        ('BenchmarkOpArrayLit_Uint8_1', 1), ('BenchmarkOpArrayLit_Uint8_10', 10),
        ('BenchmarkOpArrayLit_Uint8_100', 100), ('BenchmarkOpArrayLit_Uint8_1000', 1000)]),
    ('SliceLit', 'elements', [
        ('BenchmarkOpSliceLit_1', 1), ('BenchmarkOpSliceLit_10', 10),
        ('BenchmarkOpSliceLit_100', 100), ('BenchmarkOpSliceLit_1000', 1000)]),
    ('SliceLit2 (sparse)', 'alloc size', [
        ('BenchmarkOpSliceLit2_Sparse_10', 10), ('BenchmarkOpSliceLit2_Sparse_100', 100),
        ('BenchmarkOpSliceLit2_Sparse_1000', 1000), ('BenchmarkOpSliceLit2_Sparse_10000', 10000)]),
    ('StructLit (unnamed)', 'fields', [
        ('BenchmarkOpStructLit_1', 1), ('BenchmarkOpStructLit_10', 10),
        ('BenchmarkOpStructLit_100', 100), ('BenchmarkOpStructLit_1000', 1000)]),
    ('StructLit (named)', 'fields', [
        ('BenchmarkOpStructLitNamed_1', 1), ('BenchmarkOpStructLitNamed_10', 10),
        ('BenchmarkOpStructLitNamed_100', 100), ('BenchmarkOpStructLitNamed_1000', 1000)]),
    ('Define', 'LHS count', [
        ('BenchmarkOpDefine_1', 1), ('BenchmarkOpDefine_10', 10),
        ('BenchmarkOpDefine_100', 100), ('BenchmarkOpDefine_1000', 1000)]),
    ('Assign', 'LHS count', [
        ('BenchmarkOpAssign_1', 1), ('BenchmarkOpAssign_10', 10),
        ('BenchmarkOpAssign_100', 100), ('BenchmarkOpAssign_1000', 1000)]),
    ('Call (params)', 'params', [
        ('BenchmarkOpCall_0Params_0Captures', 0), ('BenchmarkOpCall_1Params_0Captures', 1),
        ('BenchmarkOpCall_10Params_0Captures', 10), ('BenchmarkOpCall_100Params_0Captures', 100),
        ('BenchmarkOpCall_1000Params_0Captures', 1000)]),
    ('Call (captures)', 'captures', [
        ('BenchmarkOpCall_0Params_0Captures', 0), ('BenchmarkOpCall_0Params_1Captures', 1),
        ('BenchmarkOpCall_0Params_10Captures', 10), ('BenchmarkOpCall_0Params_100Captures', 100),
        ('BenchmarkOpCall_0Params_1000Captures', 1000)]),
    ('FuncLit', 'captures', [
        ('BenchmarkOpFuncLit_Captures_0', 0), ('BenchmarkOpFuncLit_Captures_1', 1),
        ('BenchmarkOpFuncLit_Captures_10', 10), ('BenchmarkOpFuncLit_Captures_100', 100),
        ('BenchmarkOpFuncLit_Captures_1000', 1000)]),
    ('ForLoop (heap)', 'heap vars', [
        ('BenchmarkOpForLoop_HeapCopy_0', 0), ('BenchmarkOpForLoop_HeapCopy_1', 1),
        ('BenchmarkOpForLoop_HeapCopy_10', 10), ('BenchmarkOpForLoop_HeapCopy_100', 100),
        ('BenchmarkOpForLoop_HeapCopy_1000', 1000)]),
    ('RangeIter (array)', 'elements', [
        ('BenchmarkOpRangeIter_1', 1), ('BenchmarkOpRangeIter_10', 10),
        ('BenchmarkOpRangeIter_100', 100), ('BenchmarkOpRangeIter_1000', 1000)]),
    ('ReturnCallDefers', 'defers', [
        ('BenchmarkOpReturnCallDefers_1', 1), ('BenchmarkOpReturnCallDefers_10', 10),
        ('BenchmarkOpReturnCallDefers_100', 100), ('BenchmarkOpReturnCallDefers_1000', 1000)]),
    ('TypeSwitch (concrete)', 'clauses', [
        ('BenchmarkOpTypeSwitch_1', 1), ('BenchmarkOpTypeSwitch_10', 10),
        ('BenchmarkOpTypeSwitch_100', 100), ('BenchmarkOpTypeSwitch_1000', 1000)]),
    ('TypeSwitch (iface)', 'methods', [
        ('BenchmarkOpTypeSwitch_Interface_1', 1), ('BenchmarkOpTypeSwitch_Interface_10', 10),
        ('BenchmarkOpTypeSwitch_Interface_100', 100), ('BenchmarkOpTypeSwitch_Interface_1000', 1000)]),
    ('TypeAssert1 (iface)', 'methods', [
        ('BenchmarkOpTypeAssert1_Interface_1', 1), ('BenchmarkOpTypeAssert1_Interface_10', 10),
        ('BenchmarkOpTypeAssert1_Interface_100', 100)]),
    ('TypeAssert2 (iface)', 'methods', [
        ('BenchmarkOpTypeAssert2_Interface_Hit_1', 1), ('BenchmarkOpTypeAssert2_Interface_Hit_10', 10),
        ('BenchmarkOpTypeAssert2_Interface_Hit_100', 100)]),
    ('Selector (VPIface)', 'methods', [
        ('BenchmarkOpSelector_VPInterface_1', 1), ('BenchmarkOpSelector_VPInterface_10', 10),
        ('BenchmarkOpSelector_VPInterface_100', 100)]),
    ('Eval (NameExpr)', 'depth', [
        ('BenchmarkOpEval_NameExpr_1', 1), ('BenchmarkOpEval_NameExpr_10', 10),
        ('BenchmarkOpEval_NameExpr_100', 100)]),
    ('Convert (str->runes)', 'str len', [
        ('BenchmarkOpConvert_StringToRunes_1', 1), ('BenchmarkOpConvert_StringToRunes_10', 10),
        ('BenchmarkOpConvert_StringToRunes_100', 100), ('BenchmarkOpConvert_StringToRunes_1000', 1000)]),
    ('Convert (runes->str)', 'runes', [
        ('BenchmarkOpConvert_RunesToString_1', 1), ('BenchmarkOpConvert_RunesToString_10', 10),
        ('BenchmarkOpConvert_RunesToString_100', 100), ('BenchmarkOpConvert_RunesToString_1000', 1000)]),
    ('Eql (array int)', 'elements', [
        ('BenchmarkOpEql_Array_1', 1), ('BenchmarkOpEql_Array_10', 10),
        ('BenchmarkOpEql_Array_100', 100), ('BenchmarkOpEql_Array_1000', 1000)]),
    ('Eql (struct int)', 'fields', [
        ('BenchmarkOpEql_Struct_1', 1), ('BenchmarkOpEql_Struct_10', 10),
        ('BenchmarkOpEql_Struct_100', 100), ('BenchmarkOpEql_Struct_1000', 1000)]),
    ('Lss (string)', 'str len', [
        ('BenchmarkOpLss_String_10', 10), ('BenchmarkOpLss_String_100', 100),
        ('BenchmarkOpLss_String_1000', 1000), ('BenchmarkOpLss_String_10000', 10000)]),
    ('StructType', 'fields', [
        ('BenchmarkOpStructType_1', 1), ('BenchmarkOpStructType_10', 10),
        ('BenchmarkOpStructType_100', 100), ('BenchmarkOpStructType_1000', 1000)]),
    ('InterfaceType', 'methods', [
        ('BenchmarkOpInterfaceType_1', 1), ('BenchmarkOpInterfaceType_10', 10),
        ('BenchmarkOpInterfaceType_100', 100), ('BenchmarkOpInterfaceType_1000', 1000)]),
    # BigInt
    ('Add (BigInt)', 'bits', [
        ('BenchmarkOpAdd_BigInt_64', 64), ('BenchmarkOpAdd_BigInt_256', 256),
        ('BenchmarkOpAdd_BigInt_1024', 1024), ('BenchmarkOpAdd_BigInt_4096', 4096)]),
    ('Sub (BigInt)', 'bits', [
        ('BenchmarkOpSub_BigInt_64', 64), ('BenchmarkOpSub_BigInt_256', 256),
        ('BenchmarkOpSub_BigInt_1024', 1024), ('BenchmarkOpSub_BigInt_4096', 4096)]),
    ('Mul (BigInt)', 'bits', [
        ('BenchmarkOpMul_BigInt_64', 64), ('BenchmarkOpMul_BigInt_256', 256),
        ('BenchmarkOpMul_BigInt_1024', 1024), ('BenchmarkOpMul_BigInt_4096', 4096)]),
    ('Eql (BigInt)', 'bits', [
        ('BenchmarkOpEql_BigInt_64', 64), ('BenchmarkOpEql_BigInt_256', 256),
        ('BenchmarkOpEql_BigInt_1024', 1024), ('BenchmarkOpEql_BigInt_4096', 4096)]),
    ('Lss (BigInt)', 'bits', [
        ('BenchmarkOpLss_BigInt_64', 64), ('BenchmarkOpLss_BigInt_256', 256),
        ('BenchmarkOpLss_BigInt_1024', 1024), ('BenchmarkOpLss_BigInt_4096', 4096)]),
    ('Band (BigInt)', 'bits', [
        ('BenchmarkOpBand_BigInt_64', 64), ('BenchmarkOpBand_BigInt_256', 256),
        ('BenchmarkOpBand_BigInt_1024', 1024), ('BenchmarkOpBand_BigInt_4096', 4096)]),
    ('Inc (BigInt)', 'bits', [
        ('BenchmarkOpInc_BigInt_64', 64), ('BenchmarkOpInc_BigInt_256', 256),
        ('BenchmarkOpInc_BigInt_1024', 1024), ('BenchmarkOpInc_BigInt_4096', 4096)]),
    # BigDec
    ('Add (BigDec)', 'digits', [
        ('BenchmarkOpAdd_BigDec_10', 10), ('BenchmarkOpAdd_BigDec_100', 100),
        ('BenchmarkOpAdd_BigDec_1000', 1000), ('BenchmarkOpAdd_BigDec_10000', 10000)]),
    ('Sub (BigDec)', 'digits', [
        ('BenchmarkOpSub_BigDec_10', 10), ('BenchmarkOpSub_BigDec_100', 100),
        ('BenchmarkOpSub_BigDec_1000', 1000), ('BenchmarkOpSub_BigDec_10000', 10000)]),
    ('Mul (BigDec)', 'digits', [
        ('BenchmarkOpMul_BigDec_10', 10), ('BenchmarkOpMul_BigDec_100', 100),
        ('BenchmarkOpMul_BigDec_1000', 1000), ('BenchmarkOpMul_BigDec_10000', 10000)]),
    ('Quo (BigDec)', 'digits', [
        ('BenchmarkOpQuo_BigDec_10', 10), ('BenchmarkOpQuo_BigDec_100', 100),
        ('BenchmarkOpQuo_BigDec_1000', 1000), ('BenchmarkOpQuo_BigDec_10000', 10000)]),
    ('Inc (BigDec)', 'digits', [
        ('BenchmarkOpInc_BigDec_10', 10), ('BenchmarkOpInc_BigDec_100', 100),
        ('BenchmarkOpInc_BigDec_1000', 1000), ('BenchmarkOpInc_BigDec_10000', 10000)]),
]

# Ops with parameterized benchmarks that are confirmed flat for CPU gas.
# Shown separately to demonstrate they don't scale.
FLAT_FAMILIES = [
    ('RangeIterString', 'string len', [
        ('BenchmarkOpRangeIterString_1', 1), ('BenchmarkOpRangeIterString_10', 10),
        ('BenchmarkOpRangeIterString_100', 100), ('BenchmarkOpRangeIterString_1000', 1000)]),
    ('RangeIterMap', 'entries', [
        ('BenchmarkOpRangeIterMap_1', 1), ('BenchmarkOpRangeIterMap_10', 10),
        ('BenchmarkOpRangeIterMap_100', 100), ('BenchmarkOpRangeIterMap_1000', 1000)]),
    ('Defer', 'args', [
        ('BenchmarkOpDefer_1Arg', 1), ('BenchmarkOpDefer_10Args', 10),
        ('BenchmarkOpDefer_100Args', 100)]),
    ('Selector (field)', 'fields', [
        ('BenchmarkOpSelector_1field', 1), ('BenchmarkOpSelector_10fields', 10),
        ('BenchmarkOpSelector_100fields', 100), ('BenchmarkOpSelector_1000fields', 1000)]),
    ('Add (string)', 'total chars', [
        ('BenchmarkOpAdd_String_10', 10), ('BenchmarkOpAdd_String_100', 100),
        ('BenchmarkOpAdd_String_1000', 1000), ('BenchmarkOpAdd_String_10000', 10000)]),
    ('Slice (string)', 'string len', [
        ('BenchmarkOpSlice_String_10', 10), ('BenchmarkOpSlice_String_100', 100),
        ('BenchmarkOpSlice_String_1000', 1000), ('BenchmarkOpSlice_String_10000', 10000)]),
    ('Convert (str→bytes)', 'byte count', [
        ('BenchmarkOpConvert_StringToBytes_10', 10), ('BenchmarkOpConvert_StringToBytes_100', 100),
        ('BenchmarkOpConvert_StringToBytes_1000', 1000), ('BenchmarkOpConvert_StringToBytes_10000', 10000)]),
    ('Eql (string)', 'string len', [
        ('BenchmarkOpEql_String_10', 10), ('BenchmarkOpEql_String_100', 100),
        ('BenchmarkOpEql_String_1000', 1000), ('BenchmarkOpEql_String_10000', 10000)]),
    ('Eql (byte array)', 'bytes', [
        ('BenchmarkOpEql_ByteArray_1', 1), ('BenchmarkOpEql_ByteArray_10', 10),
        ('BenchmarkOpEql_ByteArray_100', 100), ('BenchmarkOpEql_ByteArray_1000', 1000)]),
]

def main():
    import sys
    import os
    script_dir = os.path.dirname(os.path.abspath(__file__))
    bench_path = sys.argv[1] if len(sys.argv) > 1 else os.path.join(script_dir, 'op_bench_do_dedicated.txt')
    out_path = sys.argv[2] if len(sys.argv) > 2 else os.path.join(script_dir, 'op_gas_fits.png')

    data = parse_benchmarks(bench_path)

    def collect(families):
        result = []
        for name, param, benchmarks in families:
            points = []
            for bname, n in benchmarks:
                s = get_stats(data, bname)
                if s:
                    points.append((n, cpu_gas(s[0], s[1])))
            if len(points) >= 2:
                result.append((name, param, points))
        return result

    param_plots = collect(FAMILIES)
    flat_plots = collect(FLAT_FAMILIES)

    ncols = 6
    param_rows = math.ceil(len(param_plots) / ncols)
    flat_rows = math.ceil(len(flat_plots) / ncols)
    total_rows = param_rows + 1 + flat_rows  # +1 for separator

    fig, axes = plt.subplots(total_rows, ncols,
                             figsize=(ncols * 3.5, total_rows * 2.8),
                             gridspec_kw={'height_ratios':
                                 [1]*param_rows + [0.15] + [1]*flat_rows})
    fig.suptitle('GnoVM Op Gas Calibration — CPU Gas vs Parameter Size\n(cpuBaseNs=5.2, Xeon 8168)',
                 fontsize=14, fontweight='bold', y=0.995)

    axes_flat_arr = axes.flatten()

    def plot_one(ax, name, param, points, is_flat_section=False):
        xs = [p[0] for p in points]
        ys = [p[1] for p in points]

        is_quad = name in QUADRATIC_FAMILIES

        if is_quad:
            quad_points = [(x*x, y) for x, y in points]
            a, b, r2 = least_squares(quad_points)
        else:
            a, b, r2 = least_squares(points)

        x_range = max(xs) / min(xs) if min(xs) > 0 else 1
        use_log = x_range > 50

        ax.scatter(xs, ys, color='#2196F3', s=30, zorder=5, edgecolors='black', linewidth=0.5)

        if use_log:
            x_fit = np.geomspace(min(xs), max(xs), 100)
        else:
            x_fit = np.linspace(min(xs), max(xs), 100)

        if is_flat_section:
            pos_ys = [y for y in ys if y >= 0]
            # Drop last value if it's a low outlier (alloc gas starting to dominate)
            if len(pos_ys) >= 3:
                rest_mean = sum(pos_ys[:-1]) / len(pos_ys[:-1])
                if pos_ys[-1] < rest_mean * 0.5:
                    pos_ys = pos_ys[:-1]
            y_mean = sum(pos_ys) / len(pos_ys) if pos_ys else 0
            y_fit = np.full_like(x_fit, y_mean)
            eq = f'flat ~{y_mean:.0f}'
            color = '#9E9E9E'
        elif is_quad:
            y_fit = a + b * x_fit * x_fit
            eq = f'{a:.0f} + {b:.4f}*N\u00b2'
            color = '#4CAF50' if r2 >= 0.95 else '#FF9800' if r2 >= 0.8 else '#F44336'
        else:
            y_fit = a + b * x_fit
            eq = f'{a:.0f} + {b:.2f}*N'
            color = '#4CAF50' if r2 >= 0.95 else '#FF9800' if r2 >= 0.8 else '#F44336'
        ax.plot(x_fit, y_fit, color=color, linewidth=1.5, alpha=0.8)

        if use_log:
            ax.set_xscale('log')
        ax.set_title(name, fontsize=8, fontweight='bold', pad=3)
        ax.text(0.05, 0.95, eq, transform=ax.transAxes,
                fontsize=6, verticalalignment='top',
                bbox=dict(boxstyle='round,pad=0.3', facecolor='white', alpha=0.8))
        ax.set_xlabel(param, fontsize=6)
        ax.set_ylabel('CPU gas', fontsize=6)
        ax.tick_params(labelsize=5)

        # For flat section, start Y from 0 unless there are negative points
        if is_flat_section and min(ys) >= 0:
            ax.set_ylim(bottom=0)

        if max(xs) > 5000 and not use_log:
            ax.ticklabel_format(axis='x', style='sci', scilimits=(0, 0))

    # Plot parameterized ops
    for i, (name, param, points) in enumerate(param_plots):
        row, col = divmod(i, ncols)
        ax = axes[row][col]
        plot_one(ax, name, param, points)
    # Hide unused in param section
    for j in range(len(param_plots), param_rows * ncols):
        row, col = divmod(j, ncols)
        axes[row][col].set_visible(False)

    # Separator row — draw after tight_layout so positions are final
    sep_row = param_rows
    for col in range(ncols):
        axes[sep_row][col].set_visible(False)

    # Plot flat ops
    for i, (name, param, points) in enumerate(flat_plots):
        row, col = divmod(i, ncols)
        ax = axes[sep_row + 1 + row][col]
        plot_one(ax, name, param, points, is_flat_section=True)
    # Hide unused in flat section
    for j in range(len(flat_plots), flat_rows * ncols):
        row, col = divmod(j, ncols)
        axes[sep_row + 1 + row][col].set_visible(False)

    plt.tight_layout(rect=[0, 0, 1, 0.97])

    # Draw separator text between parameterized and flat sections (after layout)
    sep_pos = axes[sep_row][0].get_position()
    y_mid = sep_pos.y0 + sep_pos.height / 2
    fig.text(0.5, y_mid + 0.005,
             '── Confirmed flat for CPU gas (alloc gas or per-element dispatch covers scaling) ──',
             ha='center', va='center', fontsize=10, color='#666666')
    fig.text(0.5, y_mid - 0.012,
             'Negative values = alloc gas already exceeds total gas cost at large N',
             ha='center', va='center', fontsize=7, color='#999999', style='italic')

    plt.savefig(out_path, dpi=180, bbox_inches='tight')
    print(f'Saved to {out_path}')

if __name__ == '__main__':
    main()
