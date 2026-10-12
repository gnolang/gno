// primitives.ts — PrimitiveType enum values and N-field decoder.
//
// GnoVM PrimitiveType uses `1 << iota` starting from InvalidType.
// The N field in TypedValue is an 8-byte little-endian value, base64-encoded.

export const PrimitiveTypes = {
  Invalid:      1,
  UntypedBool:  2,
  Bool:         4,
  UntypedString: 8,
  String:       16,
  Int:          32,
  Int8:         64,
  Int16:        128,
  UntypedRune:  256,
  Int32:        512,
  Int64:        1024,
  Uint:         2048,
  Uint8:        4096,
  DataByte:     8192,
  Uint16:       16384,
  Uint32:       32768,
  Uint64:       65536,
  Float32:      131072,
  Float64:      262144,
  UntypedBigint: 524288,
  UntypedBigdec: 1048576,
} as const;

export type PrimitiveTypeValue = typeof PrimitiveTypes[keyof typeof PrimitiveTypes];

const primNames: Record<number, string> = {
  [PrimitiveTypes.Bool]: "bool",
  [PrimitiveTypes.UntypedBool]: "untyped bool",
  [PrimitiveTypes.String]: "string",
  [PrimitiveTypes.UntypedString]: "untyped string",
  [PrimitiveTypes.Int]: "int",
  [PrimitiveTypes.Int8]: "int8",
  [PrimitiveTypes.Int16]: "int16",
  [PrimitiveTypes.UntypedRune]: "rune",
  [PrimitiveTypes.Int32]: "int32",
  [PrimitiveTypes.Int64]: "int64",
  [PrimitiveTypes.Uint]: "uint",
  [PrimitiveTypes.Uint8]: "uint8",
  [PrimitiveTypes.DataByte]: "databyte",
  [PrimitiveTypes.Uint16]: "uint16",
  [PrimitiveTypes.Uint32]: "uint32",
  [PrimitiveTypes.Uint64]: "uint64",
  [PrimitiveTypes.Float32]: "float32",
  [PrimitiveTypes.Float64]: "float64",
  [PrimitiveTypes.UntypedBigint]: "untyped bigint",
  [PrimitiveTypes.UntypedBigdec]: "untyped bigdec",
};

/** Returns the Go-style name for a PrimitiveType numeric value. */
export function primitiveTypeName(value: number): string {
  return primNames[value] ?? `prim(${value})`;
}

/** True if this primitive stores its value in the N field (not V). */
export function isNFieldPrimitive(value: number): boolean {
  return value !== PrimitiveTypes.String
    && value !== PrimitiveTypes.UntypedString
    && value !== PrimitiveTypes.UntypedBigint
    && value !== PrimitiveTypes.UntypedBigdec;
}

/** True if this is a signed integer type. */
export function isSignedInt(value: number): boolean {
  return value === PrimitiveTypes.Int
    || value === PrimitiveTypes.Int8
    || value === PrimitiveTypes.Int16
    || value === PrimitiveTypes.Int32
    || value === PrimitiveTypes.Int64
    || value === PrimitiveTypes.UntypedRune;
}

/**
 * Decode the base64-encoded N field into a human-readable string.
 * N is an 8-byte little-endian value.
 */
export function decodeN(base64: string, primValue: number): string {
  const raw = atob(base64);
  const buf = new Uint8Array(8);
  for (let i = 0; i < raw.length && i < 8; i++) {
    buf[i] = raw.charCodeAt(i);
  }
  const view = new DataView(buf.buffer);

  switch (primValue) {
    case PrimitiveTypes.Bool:
    case PrimitiveTypes.UntypedBool:
      return buf[0] !== 0 ? "true" : "false";

    case PrimitiveTypes.Int:
    case PrimitiveTypes.Int64:
      return view.getBigInt64(0, true).toString();

    case PrimitiveTypes.Int8:
      return view.getInt8(0).toString();

    case PrimitiveTypes.Int16:
      return view.getInt16(0, true).toString();

    case PrimitiveTypes.UntypedRune:
    case PrimitiveTypes.Int32:
      return view.getInt32(0, true).toString();

    case PrimitiveTypes.Uint:
    case PrimitiveTypes.Uint64:
      return view.getBigUint64(0, true).toString();

    case PrimitiveTypes.Uint8:
    case PrimitiveTypes.DataByte:
      return buf[0].toString();

    case PrimitiveTypes.Uint16:
      return view.getUint16(0, true).toString();

    case PrimitiveTypes.Uint32:
      return view.getUint32(0, true).toString();

    case PrimitiveTypes.Float32: {
      // N stores the raw bits as uint32 LE
      const bits = view.getUint32(0, true);
      const f32buf = new DataView(new ArrayBuffer(4));
      f32buf.setUint32(0, bits);
      return f32buf.getFloat32(0).toString();
    }

    case PrimitiveTypes.Float64:
      return view.getFloat64(0, true).toString();

    default:
      return `<raw:${base64}>`;
  }
}
