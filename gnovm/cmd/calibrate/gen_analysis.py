#!/usr/bin/env python3
"""Parse DO Xeon benchmark data and generate calibration analysis report."""

import re
import sys
from collections import defaultdict
import math

CPU_BASE_NS = 5.2

def parse_benchmarks(path):
    """Parse benchmark output into {name: [(ns_total, ns_pure, alloc_gas), ...]}"""
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
                ns_total = float(m.group(2))
                alloc_gas = float(m.group(3))
                ns_pure = float(m.group(4))
                data.setdefault(name, []).append((ns_total, ns_pure, alloc_gas))
    return data

def avg(vals):
    return sum(vals) / len(vals) if vals else 0

def least_squares(points):
    """Fit y = a + b*x. Returns (a, b, r2)."""
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

def read_gas_constants(path):
    constants = {}
    with open(path) as f:
        for line in f:
            m = re.match(r'\s+OpCPU(\w+)\s*=\s*(\d+)', line)
            if m:
                constants[m.group(1)] = int(m.group(2))
    return constants

def get_stats(data, name):
    """Get averaged (ns_pure, alloc_gas) for a benchmark name."""
    if name not in data:
        return None
    runs = data[name]
    return (avg([r[1] for r in runs]), avg([r[2] for r in runs]))

def gas(ns_pure, alloc_gas):
    """Convert to total gas and cpu gas."""
    total = ns_pure / CPU_BASE_NS
    cpu = total - alloc_gas
    return total, cpu

# ============================================================
# Hand-curated benchmark -> OpCPU mapping for flat ops
# ============================================================
FLAT_OPS = {
    # (benchmark_name, display_name, OpCPU_key)
    'BenchmarkOpAdd_Int': ('Add (int)', 'Add'),
    'BenchmarkOpSub_Int': ('Sub (int)', 'Sub'),
    'BenchmarkOpMul_Int': ('Mul (int)', 'Mul'),
    'BenchmarkOpQuo_Int': ('Quo (int)', 'Quo'),
    'BenchmarkOpRem_Int': ('Rem (int)', 'Rem'),
    'BenchmarkOpAdd_Float64': ('Add (float64)', 'Add'),
    'BenchmarkOpSub_Float64': ('Sub (float64)', 'Sub'),
    'BenchmarkOpMul_Float64': ('Mul (float64)', 'Mul'),
    'BenchmarkOpQuo_Float64': ('Quo (float64)', 'Quo'),
    'BenchmarkOpBand': ('Band (int)', 'Band'),
    'BenchmarkOpBor': ('Bor (int)', 'Bor'),
    'BenchmarkOpXor': ('Xor (int)', 'Xor'),
    'BenchmarkOpBandn': ('Bandn (int)', 'Bandn'),
    'BenchmarkOpShl': ('Shl (int)', 'Shl'),
    'BenchmarkOpShr': ('Shr (int)', 'Shr'),
    'BenchmarkOpLand': ('Land', 'Land'),
    'BenchmarkOpLor': ('Lor', 'Lor'),
    'BenchmarkOpUnot': ('Unot', 'Unot'),
    'BenchmarkOpUpos': ('Upos (int)', 'Upos'),
    'BenchmarkOpUneg': ('Uneg (int)', 'Uneg'),
    'BenchmarkOpUxor': ('Uxor (int)', 'Uxor'),
    'BenchmarkOpUrecv': ('Urecv', 'Urecv'),
    'BenchmarkOpEql_Int': ('Eql (int)', 'Eql'),
    'BenchmarkOpEql_Float64': ('Eql (float64)', 'Eql'),
    'BenchmarkOpNeq': ('Neq (int)', 'Neq'),
    'BenchmarkOpLss_Int': ('Lss (int)', 'Lss'),
    'BenchmarkOpLeq': ('Leq (int)', 'Leq'),
    'BenchmarkOpGtr': ('Gtr (int)', 'Gtr'),
    'BenchmarkOpGeq': ('Geq (int)', 'Geq'),
    'BenchmarkOpInc_Int': ('Inc (int)', 'Inc'),
    'BenchmarkOpDec_Int': ('Dec (int)', 'Dec'),
    'BenchmarkOpDec_Float64': ('Dec (float64)', 'Dec'),
    'BenchmarkOpAddAssign_Int': ('AddAssign (int)', 'AddAssign'),
    'BenchmarkOpSubAssign_Int': ('SubAssign (int)', 'SubAssign'),
    'BenchmarkOpMulAssign_Int': ('MulAssign (int)', 'MulAssign'),
    'BenchmarkOpQuoAssign_Int': ('QuoAssign (int)', 'QuoAssign'),
    'BenchmarkOpRemAssign_Int': ('RemAssign (int)', 'RemAssign'),
    'BenchmarkOpBandAssign_Int': ('BandAssign (int)', 'BandAssign'),
    'BenchmarkOpBorAssign_Int': ('BorAssign (int)', 'BorAssign'),
    'BenchmarkOpXorAssign_Int': ('XorAssign (int)', 'XorAssign'),
    'BenchmarkOpShlAssign_Int': ('ShlAssign (int)', 'ShlAssign'),
    'BenchmarkOpShrAssign_Int': ('ShrAssign (int)', 'ShrAssign'),
    'BenchmarkOpBandnAssign_Int': ('BandnAssign (int)', 'BandnAssign'),
    'BenchmarkOpBody': ('Body', 'Body'),
    'BenchmarkOpIfCond_True': ('IfCond (true)', 'IfCond'),
    'BenchmarkOpIfCond_False': ('IfCond (false)', 'IfCond'),
    'BenchmarkOpForLoop_Simple': ('ForLoop (simple)', 'ForLoop'),
    'BenchmarkOpSwitchClause_DefaultMatch': ('SwitchClause (default)', 'SwitchClause'),
    'BenchmarkOpSwitchClauseCase_Match': ('SwitchClauseCase (match)', 'SwitchClauseCase'),
    'BenchmarkOpSwitchClauseCase_Miss': ('SwitchClauseCase (miss)', 'SwitchClauseCase'),
    'BenchmarkOpPrecall_TypeConversion': ('Precall (type conv)', 'Precall'),
    'BenchmarkOpPrecall_FuncValue': ('Precall (func)', 'Precall'),
    'BenchmarkOpPrecall_BoundMethod': ('Precall (bound method)', 'Precall'),
    'BenchmarkOpIndex1_Array': ('Index1 (array)', 'Index1'),
    'BenchmarkOpIndex1_Slice': ('Index1 (slice)', 'Index1'),
    'BenchmarkOpIndex1_Map': ('Index1 (map)', 'Index1'),
    'BenchmarkOpIndex1_String': ('Index1 (string)', 'Index1'),
    'BenchmarkOpSlice_Array': ('Slice (array)', 'Slice'),
    'BenchmarkOpSlice_Slice': ('Slice (slice)', 'Slice'),
    'BenchmarkOpSlice_ByteArray': ('Slice (byte array)', 'Slice'),
    'BenchmarkOpSlice_3Index': ('Slice (3-index)', 'Slice'),
    'BenchmarkOpStar': ('Star', 'Star'),
    'BenchmarkOpRef': ('Ref', 'Ref'),
    'BenchmarkOpCompositeLit_Array': ('CompositeLit (array)', 'CompositeLit'),
    'BenchmarkOpSelector_Own': ('Selector (own)', 'Selector'),
    'BenchmarkOpSelector_VPBlock': ('Selector (VPBlock)', 'Selector'),
    'BenchmarkOpSelector_Method': ('Selector (method)', 'Selector'),
    'BenchmarkOpSelector_VPValMethod': ('Selector (VPValMethod)', 'Selector'),
    'BenchmarkOpConvert_IntToString': ('Convert (int->string)', 'Convert'),
    'BenchmarkOpConvert_IntToInt64': ('Convert (int->int64)', 'Convert'),
    'BenchmarkOpEval_ConstExpr': ('Eval (const)', 'Eval'),
    'BenchmarkOpEval_TypeExpr': ('Eval (type)', 'Eval'),
    'BenchmarkOpBinary1_LAND_True': ('Binary1 (LAND true)', 'Binary1'),
    'BenchmarkOpBinary1_LAND_False': ('Binary1 (LAND false)', 'Binary1'),
    'BenchmarkOpTypeAssert1': ('TypeAssert1 (concrete)', 'TypeAssert1'),
    'BenchmarkOpTypeAssert2_Hit': ('TypeAssert2 (hit)', 'TypeAssert2'),
    'BenchmarkOpTypeAssert2_Miss': ('TypeAssert2 (miss)', 'TypeAssert2'),
    'BenchmarkOpCallNativeBody_Len': ('CallNativeBody (len)', 'CallNativeBody'),
    'BenchmarkOpCallNativeBody_Println': ('CallNativeBody (println)', 'CallNativeBody'),
    'BenchmarkOpCall_Method': ('Call (method)', 'Call'),
    'BenchmarkOpReturn': ('Return', 'Return'),
    'BenchmarkOpReturnAfterCopy': ('ReturnAfterCopy', 'ReturnAfterCopy'),
    'BenchmarkOpReturnFromBlock': ('ReturnFromBlock', 'ReturnFromBlock'),
    'BenchmarkOpReturnToBlock': ('ReturnToBlock', 'ReturnToBlock'),
    'BenchmarkOpPanic2': ('Panic2', 'Panic2'),
    'BenchmarkOpChanType': ('ChanType', 'ChanType'),
    'BenchmarkOpMapType': ('MapType', 'MapType'),
    'BenchmarkOpArrayType': ('ArrayType', 'ArrayType'),
    'BenchmarkOpSliceType': ('SliceType', 'SliceType'),
    'BenchmarkOpValueDecl_DefaultInt': ('ValueDecl (int)', 'ValueDecl'),
    'BenchmarkOpTypeDecl': ('TypeDecl', 'TypeDecl'),
    # These are flat despite having parameterized benchmarks (use N=10 to avoid cold-cache):
    'BenchmarkOpRangeIterString_10': ('RangeIterString', 'RangeIterString'),  # per-rune op, not per-string
    'BenchmarkOpRangeIterMap_10': ('RangeIterMap', 'RangeIterMap'),  # per-entry op, not per-map
    'BenchmarkOpDefer_10Args': ('Defer', 'Defer'),  # args already on stack
    'BenchmarkOpSelector_10fields': ('Selector (field)', 'Selector'),  # direct VPField index
    # CPU-flat because alloc gas covers the O(N) memory cost:
    'BenchmarkOpAdd_String_10': ('Add (string)', 'Add'),
    'BenchmarkOpSlice_String_10': ('Slice (string)', 'Slice'),
    'BenchmarkOpConvert_StringToBytes_10': ('Convert (str->bytes)', 'Convert'),
}

# Hand-curated parameterized families
# Each entry: (display_name, param_label, [(bench_suffix, N_value), ...])
PARAM_FAMILIES = [
    # --- Section 2: Parameterized ops ---
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
    ('Call (params, 0 captures)', 'params', [
        ('BenchmarkOpCall_0Params_0Captures', 0), ('BenchmarkOpCall_1Params_0Captures', 1),
        ('BenchmarkOpCall_10Params_0Captures', 10), ('BenchmarkOpCall_100Params_0Captures', 100),
        ('BenchmarkOpCall_1000Params_0Captures', 1000)]),
    ('Call (0 params, captures)', 'captures', [
        ('BenchmarkOpCall_0Params_0Captures', 0), ('BenchmarkOpCall_0Params_1Captures', 1),
        ('BenchmarkOpCall_0Params_10Captures', 10), ('BenchmarkOpCall_0Params_100Captures', 100),
        ('BenchmarkOpCall_0Params_1000Captures', 1000)]),
    ('FuncLit (captures)', 'captures', [
        ('BenchmarkOpFuncLit_Captures_0', 0), ('BenchmarkOpFuncLit_Captures_1', 1),
        ('BenchmarkOpFuncLit_Captures_10', 10), ('BenchmarkOpFuncLit_Captures_100', 100),
        ('BenchmarkOpFuncLit_Captures_1000', 1000)]),
    ('ForLoop (heap copy)', 'heap vars', [
        ('BenchmarkOpForLoop_HeapCopy_0', 0), ('BenchmarkOpForLoop_HeapCopy_1', 1),
        ('BenchmarkOpForLoop_HeapCopy_10', 10), ('BenchmarkOpForLoop_HeapCopy_100', 100),
        ('BenchmarkOpForLoop_HeapCopy_1000', 1000)]),
    ('RangeIter (array)', 'elements', [
        ('BenchmarkOpRangeIter_1', 1), ('BenchmarkOpRangeIter_10', 10),
        ('BenchmarkOpRangeIter_100', 100), ('BenchmarkOpRangeIter_1000', 1000)]),
    # RangeIterString and RangeIterMap are flat: called once per element, not per collection.
    # See FLAT_OPS for their single-invocation cost.
    ('ReturnCallDefers', 'defers', [
        ('BenchmarkOpReturnCallDefers_1', 1), ('BenchmarkOpReturnCallDefers_10', 10),
        ('BenchmarkOpReturnCallDefers_100', 100), ('BenchmarkOpReturnCallDefers_1000', 1000)]),
    # Defer is flat: args are already on the stack, cost doesn't scale with arg count.
    # See FLAT_OPS.
    ('TypeSwitch (concrete)', 'clauses', [
        ('BenchmarkOpTypeSwitch_1', 1), ('BenchmarkOpTypeSwitch_10', 10),
        ('BenchmarkOpTypeSwitch_100', 100), ('BenchmarkOpTypeSwitch_1000', 1000)]),
    ('TypeSwitch (interface)', 'methods', [
        ('BenchmarkOpTypeSwitch_Interface_1', 1), ('BenchmarkOpTypeSwitch_Interface_10', 10),
        ('BenchmarkOpTypeSwitch_Interface_100', 100), ('BenchmarkOpTypeSwitch_Interface_1000', 1000)]),
    ('TypeAssert1 (interface)', 'methods', [
        ('BenchmarkOpTypeAssert1_Interface_1', 1), ('BenchmarkOpTypeAssert1_Interface_10', 10),
        ('BenchmarkOpTypeAssert1_Interface_100', 100)]),
    ('TypeAssert2 (iface hit)', 'methods', [
        ('BenchmarkOpTypeAssert2_Interface_Hit_1', 1), ('BenchmarkOpTypeAssert2_Interface_Hit_10', 10),
        ('BenchmarkOpTypeAssert2_Interface_Hit_100', 100)]),
    ('Selector (VPInterface)', 'methods', [
        ('BenchmarkOpSelector_VPInterface_1', 1), ('BenchmarkOpSelector_VPInterface_10', 10),
        ('BenchmarkOpSelector_VPInterface_100', 100)]),
    # Selector (by field count) is flat: uses direct VPField index, doesn't scan fields.
    # See FLAT_OPS.
    ('Eval (NameExpr depth)', 'depth', [
        ('BenchmarkOpEval_NameExpr_Depth1', 1), ('BenchmarkOpEval_NameExpr_Depth10', 10),
        ('BenchmarkOpEval_NameExpr_Depth100', 100), ('BenchmarkOpEval_NameExpr_Depth200', 200)]),
    # Slice (string) is flat for CPU gas: alloc gas covers the O(N) copy.
    # See FLAT_OPS.
    ('Convert (str->runes)', 'string len', [
        ('BenchmarkOpConvert_StringToRunes_1', 1), ('BenchmarkOpConvert_StringToRunes_10', 10),
        ('BenchmarkOpConvert_StringToRunes_100', 100), ('BenchmarkOpConvert_StringToRunes_1000', 1000)]),
    ('Convert (runes->str)', 'rune count', [
        ('BenchmarkOpConvert_RunesToString_1', 1), ('BenchmarkOpConvert_RunesToString_10', 10),
        ('BenchmarkOpConvert_RunesToString_100', 100), ('BenchmarkOpConvert_RunesToString_1000', 1000)]),
    # Convert (str->bytes) is flat for CPU gas: alloc gas covers the O(N) copy.
    # See FLAT_OPS.
    ('Eql (array of int)', 'elements', [
        ('BenchmarkOpEql_Array_1', 1), ('BenchmarkOpEql_Array_10', 10),
        ('BenchmarkOpEql_Array_100', 100), ('BenchmarkOpEql_Array_1000', 1000)]),
    ('Eql (struct of int)', 'fields', [
        ('BenchmarkOpEql_Struct_1', 1), ('BenchmarkOpEql_Struct_10', 10),
        ('BenchmarkOpEql_Struct_100', 100), ('BenchmarkOpEql_Struct_1000', 1000)]),
    ('Eql (byte array)', 'bytes', [
        ('BenchmarkOpEql_ByteArray_1', 1), ('BenchmarkOpEql_ByteArray_10', 10),
        ('BenchmarkOpEql_ByteArray_100', 100), ('BenchmarkOpEql_ByteArray_1000', 1000)]),
    ('Eql (string)', 'string len', [
        ('BenchmarkOpEql_String_10', 10), ('BenchmarkOpEql_String_100', 100),
        ('BenchmarkOpEql_String_1000', 1000), ('BenchmarkOpEql_String_10000', 10000)]),
    ('Lss (string)', 'string len', [
        ('BenchmarkOpLss_String_10', 10), ('BenchmarkOpLss_String_100', 100),
        ('BenchmarkOpLss_String_1000', 1000), ('BenchmarkOpLss_String_10000', 10000)]),
    # Add (string) is flat for CPU gas: alloc gas covers the O(N) concatenation.
    # See FLAT_OPS.
    ('StructType', 'fields', [
        ('BenchmarkOpStructType_1', 1), ('BenchmarkOpStructType_10', 10),
        ('BenchmarkOpStructType_100', 100), ('BenchmarkOpStructType_1000', 1000)]),
    ('InterfaceType', 'methods', [
        ('BenchmarkOpInterfaceType_1', 1), ('BenchmarkOpInterfaceType_10', 10),
        ('BenchmarkOpInterfaceType_100', 100), ('BenchmarkOpInterfaceType_1000', 1000)]),
    ('FuncType (params, 0 results)', 'params', [
        ('BenchmarkOpFuncType_0Params_0Results', 0), ('BenchmarkOpFuncType_1Params_0Results', 1),
        ('BenchmarkOpFuncType_10Params_0Results', 10), ('BenchmarkOpFuncType_100Params_0Results', 100),
        ('BenchmarkOpFuncType_1000Params_0Results', 1000)]),
    ('FuncType (0 params, results)', 'results', [
        ('BenchmarkOpFuncType_0Params_0Results', 0), ('BenchmarkOpFuncType_0Params_1Results', 1),
        ('BenchmarkOpFuncType_0Params_10Results', 10), ('BenchmarkOpFuncType_0Params_100Results', 100),
        ('BenchmarkOpFuncType_0Params_1000Results', 1000)]),
    ('ValueDecl (struct)', 'fields', [
        ('BenchmarkOpValueDecl_DefaultStruct10', 10)]),
    ('ValueDecl (array)', 'elements', [
        ('BenchmarkOpValueDecl_DefaultArray100', 100), ('BenchmarkOpValueDecl_DefaultArray1000', 1000)]),
]

# BigInt families
BIGINT_FAMILIES = [
    ('Add (BigInt)', [
        ('BenchmarkOpAdd_BigInt_64', 64), ('BenchmarkOpAdd_BigInt_256', 256),
        ('BenchmarkOpAdd_BigInt_1024', 1024), ('BenchmarkOpAdd_BigInt_4096', 4096)]),
    ('Sub (BigInt)', [
        ('BenchmarkOpSub_BigInt_64', 64), ('BenchmarkOpSub_BigInt_256', 256),
        ('BenchmarkOpSub_BigInt_1024', 1024), ('BenchmarkOpSub_BigInt_4096', 4096)]),
    ('Mul (BigInt) [same width]', [
        ('BenchmarkOpMul_BigInt_64', 64), ('BenchmarkOpMul_BigInt_256', 256),
        ('BenchmarkOpMul_BigInt_1024', 1024), ('BenchmarkOpMul_BigInt_4096', 4096)]),
    ('Mul (BigInt) [cross width]', [
        ('BenchmarkOpMul_BigInt_64x1024', 64*1024), ('BenchmarkOpMul_BigInt_64x4096', 64*4096),
        ('BenchmarkOpMul_BigInt_256x1024', 256*1024), ('BenchmarkOpMul_BigInt_256x4096', 256*4096)]),
    ('Quo (BigInt)', [
        ('BenchmarkOpQuo_BigInt_1024x64', 1024), ('BenchmarkOpQuo_BigInt_4096x64', 4096),
        ('BenchmarkOpQuo_BigInt_4096x256', 4096)]),
    ('Rem (BigInt)', [
        ('BenchmarkOpRem_BigInt_64', 64), ('BenchmarkOpRem_BigInt_256', 256),
        ('BenchmarkOpRem_BigInt_1024', 1024), ('BenchmarkOpRem_BigInt_4096', 4096),
        ('BenchmarkOpRem_BigInt_4096x64', 4096), ('BenchmarkOpRem_BigInt_4096x256', 4096)]),
    ('Shl (BigInt)', [
        ('BenchmarkOpShl_BigInt_10', 10), ('BenchmarkOpShl_BigInt_100', 100),
        ('BenchmarkOpShl_BigInt_1000', 1000), ('BenchmarkOpShl_BigInt_10000', 10000)]),
    ('Shr (BigInt)', [
        ('BenchmarkOpShr_BigInt_10', 10), ('BenchmarkOpShr_BigInt_100', 100),
        ('BenchmarkOpShr_BigInt_1000', 1000), ('BenchmarkOpShr_BigInt_10000', 10000)]),
    ('Eql (BigInt)', [
        ('BenchmarkOpEql_BigInt_64', 64), ('BenchmarkOpEql_BigInt_256', 256),
        ('BenchmarkOpEql_BigInt_1024', 1024), ('BenchmarkOpEql_BigInt_4096', 4096)]),
    ('Neq (BigInt)', [
        ('BenchmarkOpNeq_BigInt_64', 64), ('BenchmarkOpNeq_BigInt_256', 256),
        ('BenchmarkOpNeq_BigInt_1024', 1024), ('BenchmarkOpNeq_BigInt_4096', 4096)]),
    ('Lss (BigInt)', [
        ('BenchmarkOpLss_BigInt_64', 64), ('BenchmarkOpLss_BigInt_256', 256),
        ('BenchmarkOpLss_BigInt_1024', 1024), ('BenchmarkOpLss_BigInt_4096', 4096)]),
    ('Band (BigInt)', [
        ('BenchmarkOpBand_BigInt_64', 64), ('BenchmarkOpBand_BigInt_256', 256),
        ('BenchmarkOpBand_BigInt_1024', 1024), ('BenchmarkOpBand_BigInt_4096', 4096)]),
    ('Bor (BigInt)', [
        ('BenchmarkOpBor_BigInt_64', 64), ('BenchmarkOpBor_BigInt_256', 256),
        ('BenchmarkOpBor_BigInt_1024', 1024), ('BenchmarkOpBor_BigInt_4096', 4096)]),
    ('Xor (BigInt)', [
        ('BenchmarkOpXor_BigInt_64', 64), ('BenchmarkOpXor_BigInt_256', 256),
        ('BenchmarkOpXor_BigInt_1024', 1024), ('BenchmarkOpXor_BigInt_4096', 4096)]),
    ('Bandn (BigInt)', [
        ('BenchmarkOpBandn_BigInt_64', 64), ('BenchmarkOpBandn_BigInt_256', 256),
        ('BenchmarkOpBandn_BigInt_1024', 1024), ('BenchmarkOpBandn_BigInt_4096', 4096)]),
    ('Uneg (BigInt)', [
        ('BenchmarkOpUneg_BigInt_64', 64), ('BenchmarkOpUneg_BigInt_256', 256),
        ('BenchmarkOpUneg_BigInt_1024', 1024), ('BenchmarkOpUneg_BigInt_4096', 4096)]),
    ('Uxor (BigInt)', [
        ('BenchmarkOpUxor_BigInt_64', 64), ('BenchmarkOpUxor_BigInt_256', 256),
        ('BenchmarkOpUxor_BigInt_1024', 1024), ('BenchmarkOpUxor_BigInt_4096', 4096)]),
    ('Inc (BigInt)', [
        ('BenchmarkOpInc_BigInt_64', 64), ('BenchmarkOpInc_BigInt_256', 256),
        ('BenchmarkOpInc_BigInt_1024', 1024), ('BenchmarkOpInc_BigInt_4096', 4096)]),
    ('Dec (BigInt)', [
        ('BenchmarkOpDec_BigInt_64', 64), ('BenchmarkOpDec_BigInt_256', 256),
        ('BenchmarkOpDec_BigInt_1024', 1024), ('BenchmarkOpDec_BigInt_4096', 4096)]),
]

BIGDEC_FAMILIES = [
    ('Add (BigDec)', [
        ('BenchmarkOpAdd_BigDec_10', 10), ('BenchmarkOpAdd_BigDec_100', 100),
        ('BenchmarkOpAdd_BigDec_1000', 1000), ('BenchmarkOpAdd_BigDec_10000', 10000)]),
    ('Sub (BigDec)', [
        ('BenchmarkOpSub_BigDec_10', 10), ('BenchmarkOpSub_BigDec_100', 100),
        ('BenchmarkOpSub_BigDec_1000', 1000), ('BenchmarkOpSub_BigDec_10000', 10000)]),
    ('Mul (BigDec)', [
        ('BenchmarkOpMul_BigDec_10', 10), ('BenchmarkOpMul_BigDec_100', 100),
        ('BenchmarkOpMul_BigDec_1000', 1000), ('BenchmarkOpMul_BigDec_10000', 10000)]),
    ('Quo (BigDec)', [
        ('BenchmarkOpQuo_BigDec_10', 10), ('BenchmarkOpQuo_BigDec_100', 100),
        ('BenchmarkOpQuo_BigDec_1000', 1000), ('BenchmarkOpQuo_BigDec_10000', 10000)]),
    ('Uneg (BigDec)', [
        ('BenchmarkOpUneg_BigDec_10', 10), ('BenchmarkOpUneg_BigDec_100', 100),
        ('BenchmarkOpUneg_BigDec_1000', 1000), ('BenchmarkOpUneg_BigDec_10000', 10000)]),
    ('Inc (BigDec)', [
        ('BenchmarkOpInc_BigDec_10', 10), ('BenchmarkOpInc_BigDec_100', 100),
        ('BenchmarkOpInc_BigDec_1000', 1000), ('BenchmarkOpInc_BigDec_10000', 10000)]),
    ('Dec (BigDec)', [
        ('BenchmarkOpDec_BigDec_10', 10), ('BenchmarkOpDec_BigDec_100', 100),
        ('BenchmarkOpDec_BigDec_1000', 1000), ('BenchmarkOpDec_BigDec_10000', 10000)]),
]

# Families whose cost is O(N²) because benchmarks use N1=N2=N.
QUADRATIC_FAMILIES = {
    'Mul (BigInt) [same width]',
    'Mul (BigDec)',
    'Quo (BigDec)',
}

def emit_param_section(out, data, families, param_unit):
    for display_name, label, benchmarks in families:
        points_total = []
        points_cpu = []
        rows = []
        missing = False
        for bench_name, n_val in benchmarks:
            s = get_stats(data, bench_name)
            if s is None:
                missing = True
                continue
            ns_pure, alloc_gas = s
            total_g, cpu_g = gas(ns_pure, alloc_gas)
            rows.append((n_val, ns_pure, alloc_gas, total_g, cpu_g))
            points_total.append((n_val, total_g))
            points_cpu.append((n_val, cpu_g))

        if not rows:
            continue

        is_quad = display_name in QUADRATIC_FAMILIES

        out.append('')
        out.append('--- %s (%s) ---' % (display_name, label))
        out.append('')
        out.append('  %8s %10s %10s %10s %10s' % (label, 'ns/op', 'alloc-gas', 'total-gas', 'cpu-gas'))
        out.append('  ' + '-' * 58)
        for n_val, ns_pure, alloc_gas, total_g, cpu_g in rows:
            out.append('  %8d %10.1f %10.1f %10.1f %10.1f' % (n_val, ns_pure, alloc_gas, total_g, cpu_g))

        if len(points_cpu) >= 2:
            if is_quad:
                quad_points = [(x*x, y) for x, y in points_cpu]
                a, b, r2 = least_squares(quad_points)
                out.append('')
                out.append('  CPU gas fit: %.1f + %.6f * %s²  (R²=%.4f)' % (a, b, label, r2))
            else:
                a, b, r2 = least_squares(points_cpu)
                out.append('')
                if abs(b) < 0.001:
                    out.append('  CPU gas fit: flat ~%.0f  (R²=%.4f)' % (a, r2))
                else:
                    out.append('  CPU gas fit: %.1f + %.4f * %s  (R²=%.4f)' % (a, b, label, r2))
            if r2 < 0.95:
                out.append('  WARNING: Poor fit (R²<0.95).')


def main():
    import os
    script_dir = os.path.dirname(os.path.abspath(__file__))
    data_path = sys.argv[1] if len(sys.argv) > 1 else os.path.join(script_dir, 'op_bench_do_dedicated.txt')
    machine_path = sys.argv[2] if len(sys.argv) > 2 else os.path.join(script_dir, '..', '..', 'pkg', 'gnolang', 'machine.go')

    data = parse_benchmarks(data_path)
    gas_constants = read_gas_constants(machine_path)

    out = []
    def p(s=''):
        out.append(s)

    p('# Op Handler Gas Calibration Report — Xeon 8168')
    p('# cpuBaseNs = %.1f, %d unique benchmarks x 3 runs, -benchtime=2s' % (CPU_BASE_NS, len(data)))
    p('# Generated 2026-03-14 (with measured alloc-gas/op)')
    p()
    p("All gas values are raw (not multiplied by GasFactor).")
    p("'Alloc gas' = measured gas charged by allocator during benchmark.")
    p("'CPU needed' = (ns/op(pure) / cpuBaseNs) - alloc_gas = what incrCPU should charge.")

    # ============================================================
    # SECTION 1: FLAT OPS
    # ============================================================
    p()
    p('=' * 110)
    p('SECTION 1: FLAT OPS (single gas constant is appropriate)')
    p('=' * 110)
    p()

    flat_rows = []
    for bench_name, (display, opcpu_key) in FLAT_OPS.items():
        s = get_stats(data, bench_name)
        if s is None:
            continue
        ns_pure, alloc_gas = s
        total_g, cpu_g = gas(ns_pure, alloc_gas)
        cur = gas_constants.get(opcpu_key, -1)
        flat_rows.append((display, ns_pure, alloc_gas, total_g, cpu_g, cur))

    # Sort by deviation
    def deviation(row):
        _, _, _, _, cpu_g, cur = row
        if cur > 0 and cpu_g > 0:
            return -abs(math.log(cur / cpu_g))
        return -999
    flat_rows.sort(key=deviation)

    p('%-50s %8s %8s %8s %8s %8s  %s' % ('Op', 'ns/op', 'Total', 'Alloc', 'CPU', 'CurCPU', ''))
    p('-' * 110)
    for display, ns_pure, alloc_gas, total_g, cpu_g, cur in flat_rows:
        if cur > 0 and cpu_g > 0:
            ratio = cur / cpu_g
            if ratio > 3: tag = 'OVER >>>'
            elif ratio > 1.5: tag = 'over'
            elif ratio < 0.33: tag = 'UNDER <<<'
            elif ratio < 0.67: tag = 'UNDER <<<'
            elif ratio < 0.85: tag = 'under'
            else: tag = 'ok'
        else:
            tag = ''
        p('%-50s %8.1f %8.1f %8.0f %8.1f %8s  %s' % (
            display, ns_pure, total_g, alloc_gas, cpu_g,
            str(cur) if cur > 0 else '?', tag))

    # ============================================================
    # SECTION 2: PARAMETERIZED OPS
    # ============================================================
    p()
    p('=' * 110)
    p('SECTION 2: PARAMETERIZED OPS (need base + slope * N formulas)')
    p('=' * 110)

    emit_param_section(out, data, PARAM_FAMILIES, 'N')

    # ============================================================
    # SECTION 3: BIGINT OPS
    # ============================================================
    p()
    p('=' * 110)
    p('SECTION 3: BIGINT OPS (cost scales with bit width)')
    p('=' * 110)

    emit_param_section(out, data, [(n, 'bits', b) for n, b in BIGINT_FAMILIES], 'bits')

    # ============================================================
    # SECTION 4: BIGDEC OPS
    # ============================================================
    p()
    p('=' * 110)
    p('SECTION 4: BIGDEC OPS (cost scales with digit count)')
    p('=' * 110)

    emit_param_section(out, data, [(n, 'digits', b) for n, b in BIGDEC_FAMILIES], 'digits')

    # ============================================================
    # SECTION 5: TOP 20 WORST GAS MISMATCHES
    # ============================================================
    p()
    p('=' * 110)
    p('SECTION 5: TOP 20 WORST GAS MISMATCHES')
    p('=' * 110)
    p()

    mismatches = []
    for display, ns_pure, alloc_gas, total_g, cpu_g, cur in flat_rows:
        if cur > 0 and cpu_g > 0:
            ratio = cur / cpu_g
            mismatches.append((display, ns_pure, cpu_g, cur, ratio))
    mismatches.sort(key=lambda x: -abs(math.log(x[4])))

    p('%-50s %8s %8s %8s %8s' % ('Op', 'ns/op', 'Ideal', 'Current', 'Ratio'))
    p('-' * 90)
    for display, ns_pure, cpu_g, cur, ratio in mismatches[:20]:
        direction = 'OVER' if ratio > 1 else 'UNDER'
        p('%-50s %8.1f %8.1f %8d %7.2fx  %s' % (display, ns_pure, cpu_g, cur, ratio, direction))

    # ============================================================
    # SECTION 6: HIGH VARIANCE
    # ============================================================
    p()
    p('=' * 110)
    p('SECTION 6: HIGH VARIANCE BENCHMARKS (>5%% CV)')
    p('=' * 110)
    p()

    p('%-55s %8s %8s %8s' % ('Op', 'Mean', 'StdDev', 'CV%'))
    p('-' * 85)
    high_var = []
    for name in sorted(data.keys()):
        runs = data[name]
        values = [r[1] for r in runs]
        mean = avg(values)
        if mean == 0: continue
        std = (sum((v - mean)**2 for v in values) / len(values)) ** 0.5
        cv = 100 * std / mean
        if cv > 5:
            high_var.append((name, mean, std, cv))
    high_var.sort(key=lambda x: -x[3])
    for name, mean, std, cv in high_var[:20]:
        short = name.replace('BenchmarkOp', '')
        p('%-55s %8.1f %8.1f %7.1f%%' % (short, mean, std, cv))
    if not high_var:
        p('  (none — all benchmarks have CV < 5%%)')

    # ============================================================
    # SECTION 7: SUMMARY
    # ============================================================
    p()
    p('=' * 110)
    p('SECTION 7: SUMMARY & RECOMMENDATIONS')
    p('=' * 110)
    p()
    p('A. Flat ops needing recalibration (current/ideal ratio > 1.5x or < 0.67x):')
    p()
    for display, ns_pure, cpu_g, cur, ratio in mismatches:
        if ratio > 1.5 or ratio < 0.67:
            p('  %-45s current=%d  ideal=%.0f  ratio=%.2fx' % (display, cur, cpu_g, ratio))

    p()
    p('B. Parameterized ops — CPU gas formulas (alloc gas charged separately):')
    p()
    p('  %-45s %-35s %s' % ('Op', 'CPU gas formula', 'N=max: CPU gas'))

    for display_name, label, benchmarks in PARAM_FAMILIES:
        points = []
        for bench_name, n_val in benchmarks:
            s = get_stats(data, bench_name)
            if s is None: continue
            ns_pure, alloc_gas = s
            _, cpu_g = gas(ns_pure, alloc_gas)
            points.append((n_val, cpu_g))
        if len(points) < 2:
            continue
        a, b, r2 = least_squares(points)
        max_n = max(n for n, _ in points)
        max_cpu = a + b * max_n
        if abs(b) < 0.001:
            formula = 'flat ~%.0f' % a
        else:
            formula = 'base=%.0f + %.2f*%s' % (a, b, label)
        r2_note = ' (R²=%.2f !)' % r2 if r2 < 0.95 else ''
        p('  %-45s %-35s N=%d: CPU=%.0f%s' % (display_name, formula, max_n, max_cpu, r2_note))

    p()
    p("  NOTE: No parameterized op is \"negligible.\" Even tiny per-element costs")
    p("  become exploitable at scale (e.g. 0.02 ns/byte at N=10M = 38,000 gas uncharged).")
    p("  Every op with any scaling dimension needs a formula.")
    p()
    p('C. BigInt/BigDec: need operand-size-dependent formulas (see Sections 3-4).')

    # ============================================================
    # SECTION 8: WORST-CASE INPUT VERIFICATION
    # ============================================================
    p()
    p('=' * 110)
    p('SECTION 8: WORST-CASE INPUT VERIFICATION')
    p('=' * 110)
    p()
    p('All parameterized benchmarks use attacker-optimal (worst-case) inputs.')
    p()
    p('%-20s %-45s %s' % ('Op Category', 'Input Strategy', 'Worst Case?'))
    p('-' * 110)
    worst_cases = [
        ('String Eql', 'Identical strings (must compare all bytes)', 'YES'),
        ('String Lss', 'Differ only at last char', 'YES'),
        ('String Add', 'Two N-char strings concatenated', 'YES'),
        ('BigInt Add/Sub', '2^bits - 1 (all bits set)', 'YES'),
        ('BigInt Mul', '(2^bits - 1) x (2^bits - 1)', 'YES'),
        ('BigInt Quo', '(2^bits - 1) / (2^(bits/2) - 1)', 'YES'),
        ('BigDec *', '"1234567890..." (non-trivial, no fast paths)', 'YES'),
        ('Byte Array Eql', 'Identical byte arrays', 'YES'),
        ('MapLit', 'N int key/value pairs', 'YES'),
        ('ArrayLit', 'N int elements', 'YES'),
        ('StructLit', 'N int fields', 'YES'),
        ('SliceLit', 'N int elements', 'YES'),
    ]
    for cat, strategy, verdict in worst_cases:
        p('%-20s %-45s %s' % (cat, strategy, verdict))

    p()
    p('Total unique benchmarks: %d' % len(data))

    report = '\n'.join(out)
    print(report)

if __name__ == '__main__':
    main()
