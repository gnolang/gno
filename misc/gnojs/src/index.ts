// @gnojs/amino — JavaScript library for decoding Gno Amino JSON.
//
// Decodes the wire format from vm/qpkg_json, vm/qobject_json, vm/qtype_json
// into clean, UI-friendly StateNode trees.

export { PrimitiveTypes, primitiveTypeName, decodeN } from "./primitives.js";
export { typeName, typeKind, baseType, structFieldNames, getTypeId, isPrimitive } from "./type-utils.js";
export { decodePkg, decodeObject, decodeTypedValue } from "./decode.js";
export type { StateNode } from "./decode.js";
export type {
  AminoTypedValue,
  AminoType,
  AminoValue,
  AminoFieldType,
  AminoPrimitiveType,
  AminoPointerType,
  AminoArrayType,
  AminoSliceType,
  AminoStructType,
  AminoMapType,
  AminoFuncType,
  AminoInterfaceType,
  AminoRefType,
  AminoDeclaredType,
  AminoTypeType,
  AminoPackageType,
  AminoChanType,
  AminoStringValue,
  AminoRefValue,
  AminoObjectInfo,
  AminoStructValue,
  AminoArrayValue,
  AminoSliceValue,
  AminoPointerValue,
  AminoMapValue,
  AminoFuncValue,
  AminoHeapItemValue,
  AminoTypeValue,
  AminoExportRefValue,
  AminoBlockValue,
  QpkgResponse,
  QobjectResponse,
  QtypeResponse,
} from "./types.js";
