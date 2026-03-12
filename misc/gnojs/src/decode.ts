// decode.ts — Decode Amino JSON TypedValues into a flat, UI-friendly format.
//
// This is the core decoder: it takes raw Amino JSON (as returned by
// vm/qpkg_json, vm/qobject_json) and produces StateNode trees that
// any UI can render (tree view, table, inspector, etc.).

import type {
  AminoTypedValue,
  AminoType,
  AminoValue,
  AminoStructValue,
  AminoArrayValue,
  AminoSliceValue,
  AminoPointerValue,
  AminoMapValue,
  AminoFuncValue,
  AminoHeapItemValue,
  AminoTypeValue,
  AminoRefValue,
  AminoExportRefValue,
  AminoBlockValue,
  AminoStringValue,
  QpkgResponse,
  QobjectResponse,
} from "./types.js";
import { PrimitiveTypes, decodeN } from "./primitives.js";
import { typeName, typeKind, baseType, structFieldNames, getTypeId } from "./type-utils.js";

// ---- Output model ----

/** A decoded node suitable for rendering in any UI. */
export interface StateNode {
  /** Display name (variable name, field name, map key, index). */
  name: string;
  /** Human-readable type name. */
  type: string;
  /** Simplified kind: struct, map, array, slice, pointer, func, primitive, nil, type, package. */
  kind: string;
  /** Display value for leaf nodes. */
  value?: string;
  /** Whether this node can be expanded (has or can fetch children). */
  expandable: boolean;
  /** Inline children (already decoded). */
  children?: StateNode[];
  /** ObjectID for lazy-loading via vm/qobject_json. */
  objectId?: string;
  /** TypeID (RefType.ID) for resolving struct field names via vm/qtype_json. */
  typeId?: string;
  /** Length for collections. */
  length?: number;
}

// ---- Decoder ----

/** Decode a qpkg_json response into StateNodes. */
export function decodePkg(data: QpkgResponse): StateNode[] {
  const nodes: StateNode[] = [];
  for (let i = 0; i < data.names.length; i++) {
    const tv = data.values[i];
    if (!tv) continue;
    nodes.push(decodeTypedValue(data.names[i], tv));
  }
  return nodes;
}

/** Decode a qobject_json response into child StateNodes. */
export function decodeObject(data: QobjectResponse): StateNode[] {
  return decodeValueChildren(data.value);
}

/** Decode a single AminoTypedValue into a StateNode. */
export function decodeTypedValue(name: string, tv: AminoTypedValue): StateNode {
  const t = tv.T;
  const v = tv.V;
  const n = tv.N;

  if (!t) {
    return { name, type: "<nil>", kind: "nil", value: "nil", expandable: false };
  }

  const tName = typeName(t);
  const kind = typeKind(t);
  const bt = baseType(t);
  const typeId = getTypeId(t);

  // ---- RefValue: persisted object reference ----
  if (v && v["@type"] === "/gno.RefValue") {
    const rv = v as AminoRefValue;
    if (rv.PkgPath) {
      return { name, type: tName, kind: "package", value: rv.PkgPath, expandable: false };
    }
    return { name, type: tName, kind, expandable: true, objectId: rv.ObjectID, typeId };
  }

  // ---- ExportRefValue: cycle-breaking reference ----
  if (v && v["@type"] === "/gno.ExportRefValue") {
    const erv = v as AminoExportRefValue;
    return { name, type: tName, kind, value: `<cycle ${erv.ObjectID}>`, expandable: false };
  }

  // ---- HeapItemValue: unwrap transparently ----
  if (v && v["@type"] === "/gno.HeapItemValue") {
    const hiv = v as AminoHeapItemValue;
    return decodeTypedValue(name, hiv.Value);
  }

  // ---- TypeValue: type definition as a value ----
  if (v && v["@type"] === "/gno.TypeValue") {
    const tv2 = v as AminoTypeValue;
    return { name, type: "type", kind: "type", value: typeName(tv2.Type), expandable: false };
  }

  // ---- Primitives ----
  if (bt && bt["@type"] === "/gno.PrimitiveType") {
    const primVal = parseInt(bt.value);
    // String: value in V
    if ((primVal === PrimitiveTypes.String || primVal === PrimitiveTypes.UntypedString)
        && v && v["@type"] === "/gno.StringValue") {
      const s = (v as AminoStringValue).value;
      const display = s.length > 256
        ? JSON.stringify(s.substring(0, 256)) + "..."
        : JSON.stringify(s);
      return { name, type: tName, kind: "primitive", value: display, expandable: false };
    }
    // Numeric/bool: value in N
    if (n) {
      return { name, type: tName, kind: "primitive", value: decodeN(n, primVal), expandable: false };
    }
    // Zero value
    if (primVal === PrimitiveTypes.Bool || primVal === PrimitiveTypes.UntypedBool) {
      return { name, type: tName, kind: "primitive", value: "false", expandable: false };
    }
    if (primVal === PrimitiveTypes.String || primVal === PrimitiveTypes.UntypedString) {
      return { name, type: tName, kind: "primitive", value: '""', expandable: false };
    }
    return { name, type: tName, kind: "primitive", value: "0", expandable: false };
  }

  // ---- Struct (inline) ----
  if (v && v["@type"] === "/gno.StructValue") {
    const sv = v as AminoStructValue;
    const objectId = sv.ObjectInfo?.ID;
    const fieldNames = structFieldNames(bt);
    const children = sv.Fields.map((ftv, i) => {
      const fname = fieldNames && i < fieldNames.length ? fieldNames[i] : String(i);
      return decodeTypedValue(fname, ftv);
    });
    return {
      name, type: tName, kind: "struct", expandable: children.length > 0,
      children, objectId, typeId, length: sv.Fields.length,
    };
  }

  // ---- Array (inline) ----
  if (v && v["@type"] === "/gno.ArrayValue") {
    const av = v as AminoArrayValue;
    const objectId = av.ObjectInfo?.ID;
    if (av.Data) {
      const len = atob(av.Data).length;
      return { name, type: tName, kind: "array", value: `[${len}]byte{...}`, expandable: false, objectId, length: len };
    }
    const list = av.List || [];
    const children = list.map((etv, i) => decodeTypedValue(String(i), etv));
    return {
      name, type: tName, kind: "array", expandable: list.length > 0,
      children, objectId, length: list.length,
    };
  }

  // ---- Slice ----
  if (v && v["@type"] === "/gno.SliceValue") {
    const sv = v as AminoSliceValue;
    const length = parseInt(sv.Length) || 0;
    // Base is a RefValue — lazy
    if (sv.Base && sv.Base["@type"] === "/gno.RefValue") {
      const rv = sv.Base as AminoRefValue;
      return { name, type: tName, kind: "slice", expandable: length > 0, objectId: rv.ObjectID, length };
    }
    // Base is inline ArrayValue
    if (sv.Base && sv.Base["@type"] === "/gno.ArrayValue") {
      const av = sv.Base as AminoArrayValue;
      const offset = parseInt(sv.Offset) || 0;
      if (av.Data) {
        return { name, type: tName, kind: "slice", value: `[]byte (len=${length})`, expandable: false, length };
      }
      const list = (av.List || []).slice(offset, offset + length);
      const children = list.map((etv, i) => decodeTypedValue(String(i), etv));
      return { name, type: tName, kind: "slice", expandable: children.length > 0, children, length };
    }
    return { name, type: tName, kind: "slice", expandable: length > 0, length };
  }

  // ---- Map (inline) ----
  if (v && v["@type"] === "/gno.MapValue") {
    const mv = v as AminoMapValue;
    const objectId = mv.ObjectInfo?.ID;
    const items = mv.List?.List || [];
    const children = items.map(item => {
      const keyStr = previewTypedValue(item.Key);
      return decodeTypedValue(keyStr, item.Value);
    });
    return {
      name, type: tName, kind: "map", expandable: children.length > 0,
      children, objectId, length: items.length,
    };
  }

  // ---- Pointer (inline) ----
  if (v && v["@type"] === "/gno.PointerValue") {
    const pv = v as AminoPointerValue;
    if (pv.Base && pv.Base["@type"] === "/gno.RefValue") {
      const rv = pv.Base as AminoRefValue;
      return { name, type: tName, kind: "pointer", expandable: true, objectId: rv.ObjectID };
    }
    if (pv.TV) {
      const child = decodeTypedValue("*", pv.TV);
      return { name, type: tName, kind: "pointer", expandable: true, children: [child] };
    }
    return { name, type: tName, kind: "pointer", value: "nil", expandable: false };
  }

  // ---- Func ----
  if (v && v["@type"] === "/gno.FuncValue") {
    const fv = v as AminoFuncValue;
    return { name, type: tName, kind: "func", value: `func ${fv.Name}()`, expandable: false };
  }

  // ---- Zero value (type but no value) ----
  if (!v && !n) {
    return { name, type: tName, kind, value: "<zero>", expandable: false };
  }

  // ---- Fallback ----
  return { name, type: tName, kind, value: v ? `<${v["@type"]}>` : "<unknown>", expandable: false };
}

// ---- Internal helpers ----

/** Decode the children of a raw Amino Value (from qobject_json). */
function decodeValueChildren(v: AminoValue): StateNode[] {
  if (!v) return [];

  switch (v["@type"]) {
    case "/gno.StructValue": {
      const sv = v as AminoStructValue;
      return sv.Fields.map((ftv, i) => decodeTypedValue(String(i), ftv));
    }
    case "/gno.ArrayValue": {
      const av = v as AminoArrayValue;
      if (av.Data) {
        return [{ name: "data", type: "[]byte", kind: "primitive", value: `[${atob(av.Data).length}]byte{...}`, expandable: false }];
      }
      return (av.List || []).map((etv, i) => decodeTypedValue(String(i), etv));
    }
    case "/gno.MapValue": {
      const mv = v as AminoMapValue;
      return (mv.List?.List || []).map(item => {
        const keyStr = previewTypedValue(item.Key);
        return decodeTypedValue(keyStr, item.Value);
      });
    }
    case "/gno.HeapItemValue": {
      const hiv = v as AminoHeapItemValue;
      return [decodeTypedValue("value", hiv.Value)];
    }
    case "/gno.Block": {
      const block = v as AminoBlockValue;
      return (block.Values || []).map((tv, i) => decodeTypedValue(String(i), tv));
    }
    default:
      return [];
  }
}

/** Short preview string for a TypedValue (used for map keys). */
function previewTypedValue(tv: AminoTypedValue): string {
  const t = tv.T;
  const v = tv.V;
  const n = tv.N;

  if (!t) return "nil";

  const bt = baseType(t);
  if (bt && bt["@type"] === "/gno.PrimitiveType") {
    const primVal = parseInt(bt.value);
    if ((primVal === PrimitiveTypes.String || primVal === PrimitiveTypes.UntypedString)
        && v && v["@type"] === "/gno.StringValue") {
      const s = (v as AminoStringValue).value;
      return s.length > 64 ? JSON.stringify(s.substring(0, 64)) + "..." : JSON.stringify(s);
    }
    if (n) return decodeN(n, primVal);
  }
  return typeName(t);
}
