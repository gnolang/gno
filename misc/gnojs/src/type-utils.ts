// type-utils.ts — Utilities for working with Amino-encoded Gno types.

import type {
  AminoType,
  AminoFieldType,
  AminoPrimitiveType,
  AminoDeclaredType,
  AminoStructType,
} from "./types.js";
import { primitiveTypeName } from "./primitives.js";

/** Returns a human-readable display name for a type. */
export function typeName(t: AminoType | undefined): string {
  if (!t) return "<nil>";
  switch (t["@type"]) {
    case "/gno.PrimitiveType":
      return primitiveTypeName(parseInt(t.value));
    case "/gno.PointerType":
      return "*" + typeName(t.Elt);
    case "/gno.ArrayType":
      return `[${t.Len}]${typeName(t.Elt)}`;
    case "/gno.SliceType":
      return "[]" + typeName(t.Elt);
    case "/gno.MapType":
      return `map[${typeName(t.Key)}]${typeName(t.Value)}`;
    case "/gno.StructType":
      return "struct{...}";
    case "/gno.FuncType":
      return "func(...)";
    case "/gno.InterfaceType":
      return "interface{...}";
    case "/gno.RefType": {
      const id = t.ID;
      const dot = id.lastIndexOf(".");
      if (dot >= 0) {
        const pkgPath = id.substring(0, dot);
        const parts = pkgPath.split("/");
        return parts[parts.length - 1] + id.substring(dot);
      }
      return id;
    }
    case "/gno.DeclaredType": {
      const parts = (t.PkgPath || "").split("/");
      return parts[parts.length - 1] + "." + t.Name;
    }
    case "/gno.TypeType":
      return "type";
    case "/gno.PackageType":
      return "package";
    case "/gno.ChanType":
      return `chan ${typeName(t.Elt)}`;
    default:
      return t["@type"].replace("/gno.", "");
  }
}

/** Returns a simplified kind string for a type. */
export function typeKind(t: AminoType | undefined): string {
  if (!t) return "nil";
  switch (t["@type"]) {
    case "/gno.PrimitiveType": return "primitive";
    case "/gno.PointerType": return "pointer";
    case "/gno.ArrayType": return "array";
    case "/gno.SliceType": return "slice";
    case "/gno.StructType": return "struct";
    case "/gno.MapType": return "map";
    case "/gno.FuncType": return "func";
    case "/gno.InterfaceType": return "interface";
    case "/gno.RefType": return "ref";
    case "/gno.DeclaredType": return typeKind((t as AminoDeclaredType).Base);
    case "/gno.TypeType": return "type";
    case "/gno.PackageType": return "package";
    case "/gno.ChanType": return "chan";
    default: return "unknown";
  }
}

/** Unwrap DeclaredType to its base type. Returns t unchanged if not declared. */
export function baseType(t: AminoType | undefined): AminoType | undefined {
  if (!t) return undefined;
  if (t["@type"] === "/gno.DeclaredType") return (t as AminoDeclaredType).Base;
  return t;
}

/** Extract field names from a StructType. */
export function structFieldNames(t: AminoType | undefined): string[] | undefined {
  if (!t) return undefined;
  if (t["@type"] === "/gno.StructType") {
    return (t as AminoStructType).Fields.map((f: AminoFieldType) => f.Name);
  }
  if (t["@type"] === "/gno.DeclaredType") {
    return structFieldNames((t as AminoDeclaredType).Base);
  }
  return undefined;
}

/** Extract the RefType ID from a type if it is one, or from a DeclaredType. */
export function getTypeId(t: AminoType | undefined): string | undefined {
  if (!t) return undefined;
  if (t["@type"] === "/gno.RefType") return t.ID;
  if (t["@type"] === "/gno.DeclaredType") {
    return (t as AminoDeclaredType).PkgPath + "." + (t as AminoDeclaredType).Name;
  }
  return undefined;
}

/** Check if a type is a PrimitiveType with the given numeric value. */
export function isPrimitive(t: AminoType | undefined): t is AminoPrimitiveType {
  return t !== undefined && t["@type"] === "/gno.PrimitiveType";
}

/** Build a human-readable function signature from a FuncType. */
export function funcSignature(t: AminoType | undefined): string {
  if (!t || t["@type"] !== "/gno.FuncType") return "func()";
  const ft = t as import("./types.js").AminoFuncType;
  const params = (ft.Params || [])
    .filter(p => !(p.Name.startsWith("cur") && p.Type?.["@type"] === "/gno.RefType"))
    .map(p => {
      const tn = typeName(p.Type);
      return p.Name && !p.Name.startsWith(".") ? `${p.Name} ${tn}` : tn;
    })
    .join(", ");
  const results = (ft.Results || [])
    .map(r => {
      const tn = typeName(r.Type);
      return r.Name && !r.Name.startsWith(".") ? `${r.Name} ${tn}` : tn;
    });
  const retStr = results.length === 0 ? ""
    : results.length === 1 ? ` ${results[0]}`
    : ` (${results.join(", ")})`;
  return `func(${params})${retStr}`;
}
