// types.ts — Amino JSON wire types as they come from vm/qpkg_json, vm/qobject_json, vm/qtype_json.

// ---- TypedValue ----

export interface AminoTypedValue {
  T?: AminoType;
  V?: AminoValue;
  N?: string; // base64-encoded 8-byte little-endian primitive value
}

// ---- Types (discriminated by @type) ----

export type AminoType =
  | AminoPrimitiveType
  | AminoPointerType
  | AminoArrayType
  | AminoSliceType
  | AminoStructType
  | AminoMapType
  | AminoFuncType
  | AminoInterfaceType
  | AminoRefType
  | AminoDeclaredType
  | AminoTypeType
  | AminoPackageType
  | AminoChanType
  | AminoUnknownType;

export interface AminoPrimitiveType {
  "@type": "/gno.PrimitiveType";
  value: string; // numeric string, e.g. "32" for IntType
}

export interface AminoPointerType {
  "@type": "/gno.PointerType";
  Elt: AminoType;
}

export interface AminoArrayType {
  "@type": "/gno.ArrayType";
  Len: number;
  Elt: AminoType;
  Vrd: boolean;
}

export interface AminoSliceType {
  "@type": "/gno.SliceType";
  Elt: AminoType;
  Vrd: boolean;
}

export interface AminoStructType {
  "@type": "/gno.StructType";
  PkgPath: string;
  Fields: AminoFieldType[];
}

export interface AminoMapType {
  "@type": "/gno.MapType";
  Key: AminoType;
  Value: AminoType;
}

export interface AminoFuncType {
  "@type": "/gno.FuncType";
  Params: AminoFieldType[];
  Results: AminoFieldType[];
}

export interface AminoInterfaceType {
  "@type": "/gno.InterfaceType";
  PkgPath: string;
  Methods: AminoFieldType[];
  Generic: string;
}

export interface AminoRefType {
  "@type": "/gno.RefType";
  ID: string; // TypeID, e.g. "gno.land/r/demo/boards.Board"
}

export interface AminoDeclaredType {
  "@type": "/gno.DeclaredType";
  PkgPath: string;
  Name: string;
  Base: AminoType;
  Methods: AminoTypedValue[];
  ParentLoc?: unknown;
}

export interface AminoTypeType {
  "@type": "/gno.TypeType";
}

export interface AminoPackageType {
  "@type": "/gno.PackageType";
}

export interface AminoChanType {
  "@type": "/gno.ChanType";
  Dir: number;
  Elt: AminoType;
}

export interface AminoUnknownType {
  "@type": string;
  [key: string]: unknown;
}

// ---- Field type ----

export interface AminoFieldType {
  Name: string;
  Type: AminoType;
  Embedded: boolean;
  Tag: string;
}

// ---- Values (discriminated by @type) ----

export type AminoValue =
  | AminoStringValue
  | AminoRefValue
  | AminoStructValue
  | AminoArrayValue
  | AminoSliceValue
  | AminoPointerValue
  | AminoMapValue
  | AminoFuncValue
  | AminoHeapItemValue
  | AminoTypeValue
  | AminoExportRefValue
  | AminoBlockValue
  | AminoUnknownValue;

export interface AminoStringValue {
  "@type": "/gno.StringValue";
  value: string;
}

export interface AminoRefValue {
  "@type": "/gno.RefValue";
  ObjectID: string;
  Hash?: string;
  PkgPath?: string;
  Escaped?: boolean;
}

export interface AminoObjectInfo {
  ID: string;
  Hash?: string;
  OwnerID?: string;
  ModTime?: string;
  RefCount?: string;
  LastObjectSize?: string;
}

export interface AminoStructValue {
  "@type": "/gno.StructValue";
  ObjectInfo?: AminoObjectInfo;
  Fields: AminoTypedValue[];
}

export interface AminoArrayValue {
  "@type": "/gno.ArrayValue";
  ObjectInfo?: AminoObjectInfo;
  List?: AminoTypedValue[];
  Data?: string; // base64 for byte arrays
}

export interface AminoSliceValue {
  "@type": "/gno.SliceValue";
  Base: AminoValue;
  Offset: string;
  Length: string;
  Maxcap: string;
}

export interface AminoPointerValue {
  "@type": "/gno.PointerValue";
  TV: AminoTypedValue | null;
  Base: AminoValue;
  Index: string;
}

export interface AminoMapList {
  List?: AminoMapItem[];
}

export interface AminoMapItem {
  Key: AminoTypedValue;
  Value: AminoTypedValue;
}

export interface AminoMapValue {
  "@type": "/gno.MapValue";
  ObjectInfo?: AminoObjectInfo;
  List: AminoMapList;
}

export interface AminoFuncValue {
  "@type": "/gno.FuncValue";
  ObjectInfo?: AminoObjectInfo;
  Name: string;
  PkgPath?: string;
  NativePkg?: string;
  NativeName?: string;
  [key: string]: unknown;
}

export interface AminoHeapItemValue {
  "@type": "/gno.HeapItemValue";
  ObjectInfo?: AminoObjectInfo;
  Value: AminoTypedValue;
}

export interface AminoTypeValue {
  "@type": "/gno.TypeValue";
  Type: AminoType;
}

export interface AminoExportRefValue {
  "@type": "/gno.ExportRefValue";
  ObjectID: string; // ":1", ":2", etc.
}

export interface AminoBlockValue {
  "@type": "/gno.Block";
  ObjectInfo?: AminoObjectInfo;
  Source?: unknown;
  Values?: AminoTypedValue[];
  Parent?: AminoValue;
}

export interface AminoUnknownValue {
  "@type": string;
  [key: string]: unknown;
}

// ---- Endpoint response types ----

export interface QpkgResponse {
  names: string[];
  values: AminoTypedValue[];
}

export interface QobjectResponse {
  objectid: string;
  value: AminoValue;
}

export interface QtypeResponse {
  typeid: string;
  type: AminoType;
}
