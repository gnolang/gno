// decode.test.ts — Tests for the Amino JSON decoder.
// Uses real output from vm/qpkg_json and vm/qobject_json Go tests.

import { decodePkg, decodeObject, decodeTypedValue } from "./decode.js";
import type { QpkgResponse, QobjectResponse } from "./types.js";

// ---- Test helpers ----

function assert(condition: boolean, msg: string): void {
  if (!condition) throw new Error(`FAIL: ${msg}`);
}

function assertEqual<T>(actual: T, expected: T, msg: string): void {
  if (actual !== expected) {
    throw new Error(`FAIL: ${msg}\n  expected: ${JSON.stringify(expected)}\n  actual:   ${JSON.stringify(actual)}`);
  }
}

// ---- Test fixtures (real Go test output) ----

const qpkgFixture: QpkgResponse = {
  names: ["MyStruct", "myInt", "myStr", "myStruct", "init.4", "Render"],
  values: [
    { T: { "@type": "/gno.TypeType" }, V: { "@type": "/gno.TypeValue", Type: { "@type": "/gno.DeclaredType", PkgPath: "gno.land/r/test/qpkg", Name: "MyStruct", Base: { "@type": "/gno.StructType", PkgPath: "gno.land/r/test/qpkg", Fields: [{ Name: "Name", Type: { "@type": "/gno.PrimitiveType", value: "16" }, Embedded: false, Tag: "" }, { Name: "Age", Type: { "@type": "/gno.PrimitiveType", value: "32" }, Embedded: false, Tag: "" }] }, Methods: [] } } },
    { T: { "@type": "/gno.PrimitiveType", value: "32" }, N: "KgAAAAAAAAA=" },
    { T: { "@type": "/gno.PrimitiveType", value: "16" }, V: { "@type": "/gno.StringValue", value: "hello" } },
    { T: { "@type": "/gno.RefType", ID: "gno.land/r/test/qpkg.MyStruct" }, V: { "@type": "/gno.RefValue", ObjectID: "715383ba05505afed61caa873216e2ee896bede9:10", Hash: "abc" } },
    { T: { "@type": "/gno.FuncType", Params: [], Results: [] }, V: { "@type": "/gno.RefValue", ObjectID: "715383ba05505afed61caa873216e2ee896bede9:7", Hash: "def" } },
    { T: { "@type": "/gno.FuncType", Params: [{ Name: "path", Type: { "@type": "/gno.PrimitiveType", value: "16" }, Embedded: false, Tag: "" }], Results: [{ Name: ".res.0", Type: { "@type": "/gno.PrimitiveType", value: "16" }, Embedded: false, Tag: "" }] }, V: { "@type": "/gno.RefValue", ObjectID: "715383ba05505afed61caa873216e2ee896bede9:9", Hash: "ghi" } },
  ],
};

// ---- Tests ----

function testDecodePkg(): void {
  const nodes = decodePkg(qpkgFixture);
  assertEqual(nodes.length, 6, "should decode 6 nodes");

  // MyStruct — TypeValue
  assertEqual(nodes[0].name, "MyStruct", "first node name");
  assertEqual(nodes[0].kind, "type", "TypeValue kind");
  assert(nodes[0].value !== undefined, "TypeValue should have value");

  // myInt — int = 42 (N: "KgAAAAAAAAA=" is base64 of 42 as int64 LE)
  assertEqual(nodes[1].name, "myInt", "second node name");
  assertEqual(nodes[1].type, "int", "myInt type");
  assertEqual(nodes[1].kind, "primitive", "myInt kind");
  assertEqual(nodes[1].value, "42", "myInt value decoded from N");

  // myStr — string = "hello"
  assertEqual(nodes[2].name, "myStr", "third node name");
  assertEqual(nodes[2].type, "string", "myStr type");
  assertEqual(nodes[2].kind, "primitive", "myStr kind");
  assertEqual(nodes[2].value, '"hello"', "myStr value");

  // myStruct — RefValue (persisted struct)
  assertEqual(nodes[3].name, "myStruct", "fourth node name");
  assertEqual(nodes[3].kind, "ref", "myStruct kind is ref (RefType)");
  assert(nodes[3].expandable, "myStruct should be expandable");
  assertEqual(nodes[3].objectId, "715383ba05505afed61caa873216e2ee896bede9:10", "myStruct objectId");
  assertEqual(nodes[3].typeId, "gno.land/r/test/qpkg.MyStruct", "myStruct typeId");

  // init.4 — func (RefValue)
  assertEqual(nodes[4].name, "init.4", "fifth node name");
  // Func values stored as RefValue are not expandable in a useful way
  // but the kind comes from FuncType

  // Render — func (RefValue)
  assertEqual(nodes[5].name, "Render", "sixth node name");

  console.log("  testDecodePkg: PASS");
}

function testDecodeObject(): void {
  // Simulate qobject_json response for a StructValue
  const fixture: QobjectResponse = {
    objectid: "abc123:8",
    value: {
      "@type": "/gno.StructValue",
      ObjectInfo: { ID: "abc123:8" },
      Fields: [
        { T: { "@type": "/gno.PrimitiveType", value: "32" }, N: "AQAAAAAAAAA=" },
        { T: { "@type": "/gno.PrimitiveType", value: "16" }, V: { "@type": "/gno.StringValue", value: "test" } },
      ],
    },
  };

  const nodes = decodeObject(fixture);
  assertEqual(nodes.length, 2, "should decode 2 fields");

  // Field 0: int = 1
  assertEqual(nodes[0].name, "0", "field names are indices from qobject");
  assertEqual(nodes[0].type, "int", "field 0 type");
  assertEqual(nodes[0].value, "1", "field 0 value");

  // Field 1: string = "test"
  assertEqual(nodes[1].name, "1", "field 1 name");
  assertEqual(nodes[1].value, '"test"', "field 1 value");

  console.log("  testDecodeObject: PASS");
}

function testDecodeHeapItem(): void {
  // HeapItemValue should be unwrapped transparently
  const node = decodeTypedValue("ptr", {
    T: { "@type": "/gno.PointerType", Elt: { "@type": "/gno.PrimitiveType", value: "32" } },
    V: {
      "@type": "/gno.HeapItemValue",
      Value: {
        T: { "@type": "/gno.PrimitiveType", value: "32" },
        N: "BwAAAAAAAAA=", // 7
      },
    },
  });

  // HeapItemValue is unwrapped, so we get the inner primitive
  assertEqual(node.name, "ptr", "heap item preserves name");
  assertEqual(node.value, "7", "heap item unwraps to inner value");
  assertEqual(node.kind, "primitive", "heap item unwraps to primitive kind");

  console.log("  testDecodeHeapItem: PASS");
}

function testDecodeMap(): void {
  const node = decodeTypedValue("myMap", {
    T: { "@type": "/gno.MapType", Key: { "@type": "/gno.PrimitiveType", value: "16" }, Value: { "@type": "/gno.PrimitiveType", value: "32" } },
    V: {
      "@type": "/gno.MapValue",
      List: {
        List: [
          {
            Key: { T: { "@type": "/gno.PrimitiveType", value: "16" }, V: { "@type": "/gno.StringValue", value: "alice" } },
            Value: { T: { "@type": "/gno.PrimitiveType", value: "32" }, N: "ZAAAAAAAAAA=" }, // 100
          },
        ],
      },
    },
  });

  assertEqual(node.kind, "map", "map kind");
  assertEqual(node.length, 1, "map length");
  assert(node.children !== undefined && node.children.length === 1, "map should have 1 child");
  assertEqual(node.children![0].name, '"alice"', "map key as child name");
  assertEqual(node.children![0].value, "100", "map value");

  console.log("  testDecodeMap: PASS");
}

function testDecodePointerRefValue(): void {
  const node = decodeTypedValue("ptr", {
    T: { "@type": "/gno.PointerType", Elt: { "@type": "/gno.RefType", ID: "gno.land/r/test.Foo" } },
    V: {
      "@type": "/gno.PointerValue",
      TV: null,
      Base: { "@type": "/gno.RefValue", ObjectID: "abc:5", Hash: "xyz" },
      Index: "0",
    },
  });

  assertEqual(node.kind, "pointer", "pointer kind");
  assert(node.expandable, "pointer with RefValue base should be expandable");
  assertEqual(node.objectId, "abc:5", "pointer objectId from Base");

  console.log("  testDecodePointerRefValue: PASS");
}

function testDecodeInlineStruct(): void {
  const node = decodeTypedValue("s", {
    T: { "@type": "/gno.StructType", PkgPath: "gno.land/r/test", Fields: [
      { Name: "X", Type: { "@type": "/gno.PrimitiveType", value: "32" }, Embedded: false, Tag: "" },
      { Name: "Y", Type: { "@type": "/gno.PrimitiveType", value: "16" }, Embedded: false, Tag: "" },
    ]},
    V: {
      "@type": "/gno.StructValue",
      Fields: [
        { T: { "@type": "/gno.PrimitiveType", value: "32" }, N: "CQAAAAAAAAA=" }, // 9
        { T: { "@type": "/gno.PrimitiveType", value: "16" }, V: { "@type": "/gno.StringValue", value: "hi" } },
      ],
    },
  });

  assertEqual(node.kind, "struct", "struct kind");
  assertEqual(node.length, 2, "struct length");
  assert(node.children !== undefined && node.children.length === 2, "struct has 2 children");
  assertEqual(node.children![0].name, "X", "field name from StructType");
  assertEqual(node.children![0].value, "9", "field X value");
  assertEqual(node.children![1].name, "Y", "field name Y");
  assertEqual(node.children![1].value, '"hi"', "field Y value");

  console.log("  testDecodeInlineStruct: PASS");
}

function testDecodeCycleRef(): void {
  const node = decodeTypedValue("self", {
    T: { "@type": "/gno.PointerType", Elt: { "@type": "/gno.RefType", ID: "gno.land/r/test.Node" } },
    V: { "@type": "/gno.ExportRefValue", ObjectID: ":1" },
  });

  assertEqual(node.kind, "pointer", "cycle ref kind");
  assertEqual(node.value, "<cycle :1>", "cycle ref value");
  assert(!node.expandable, "cycle ref should not be expandable");

  console.log("  testDecodeCycleRef: PASS");
}

function testDecodeNil(): void {
  const node = decodeTypedValue("x", {});
  assertEqual(node.kind, "nil", "nil kind");
  assertEqual(node.value, "nil", "nil value");

  console.log("  testDecodeNil: PASS");
}

function testDecodeSliceRefBase(): void {
  const node = decodeTypedValue("items", {
    T: { "@type": "/gno.SliceType", Elt: { "@type": "/gno.PrimitiveType", value: "32" }, Vrd: false },
    V: {
      "@type": "/gno.SliceValue",
      Base: { "@type": "/gno.RefValue", ObjectID: "abc:3", Hash: "def" },
      Offset: "0",
      Length: "5",
      Maxcap: "8",
    },
  });

  assertEqual(node.kind, "slice", "slice kind");
  assertEqual(node.length, 5, "slice length");
  assert(node.expandable, "slice with RefValue base should be expandable");
  assertEqual(node.objectId, "abc:3", "slice objectId from Base");

  console.log("  testDecodeSliceRefBase: PASS");
}

function testDecodeFuncInline(): void {
  const node = decodeTypedValue("myFunc", {
    T: { "@type": "/gno.FuncType", Params: [{ Name: "x", Type: { "@type": "/gno.PrimitiveType", value: "32" }, Embedded: false, Tag: "" }], Results: [{ Name: ".res.0", Type: { "@type": "/gno.PrimitiveType", value: "32" }, Embedded: false, Tag: "" }] },
    V: {
      "@type": "/gno.FuncValue",
      Type: { "@type": "/gno.FuncType", Params: [{ Name: "x", Type: { "@type": "/gno.PrimitiveType", value: "32" }, Embedded: false, Tag: "" }], Results: [{ Name: ".res.0", Type: { "@type": "/gno.PrimitiveType", value: "32" }, Embedded: false, Tag: "" }] },
      Name: "myFunc",
      Source: { "@type": "/gno.RefNode", Location: { PkgPath: "gno.land/r/test", File: "test.gno", Span: { Pos: { Line: "5", Column: "1" }, End: { Line: "7", Column: "1" }, Num: "0" } }, BlockNode: null },
    },
  });

  assertEqual(node.kind, "func", "inline func kind");
  assert(node.source !== undefined, "inline func should have source");
  assertEqual(node.source!.file, "test.gno", "func source file");
  assertEqual(node.source!.startLine, 5, "func source start line");

  console.log("  testDecodeFuncInline: PASS");
}

function testDecodeFuncRefValue(): void {
  // Func stored as RefValue (top-level package func) — expandable to show source
  const node = decodeTypedValue("Render", {
    T: { "@type": "/gno.FuncType", Params: [{ Name: "path", Type: { "@type": "/gno.PrimitiveType", value: "16" }, Embedded: false, Tag: "" }], Results: [{ Name: ".res.0", Type: { "@type": "/gno.PrimitiveType", value: "16" }, Embedded: false, Tag: "" }] },
    V: { "@type": "/gno.RefValue", ObjectID: "abc:9", Hash: "xyz" },
  });

  assertEqual(node.kind, "func", "func ref kind");
  assert(node.expandable, "func ref should be expandable");
  assertEqual(node.objectId, "abc:9", "func ref objectId");
  assert(node.type.includes("func("), "func type should include signature");

  console.log("  testDecodeFuncRefValue: PASS");
}

function testDecodeClosureWithCaptures(): void {
  // Closure with captures — the FuncValue has Captures field
  const node = decodeTypedValue("stepper", {
    T: { "@type": "/gno.FuncType", Params: [], Results: [{ Name: ".res.0", Type: { "@type": "/gno.PrimitiveType", value: "32" }, Embedded: false, Tag: "" }] },
    V: {
      "@type": "/gno.FuncValue",
      Type: { "@type": "/gno.FuncType", Params: [], Results: [{ Name: ".res.0", Type: { "@type": "/gno.PrimitiveType", value: "32" }, Embedded: false, Tag: "" }] },
      Name: "",
      IsClosure: false,
      Captures: [
        { T: { "@type": "/gno.heapItemType" }, V: { "@type": "/gno.RefValue", ObjectID: "abc:13", Hash: "def" } },
      ],
      Source: { "@type": "/gno.RefNode", Location: { PkgPath: "gno.land/r/test", File: "test.gno", Span: { Pos: { Line: "17", Column: "12" }, End: { Line: "20", Column: "3" }, Num: "0" } }, BlockNode: null },
    },
  });

  assertEqual(node.kind, "closure", "closure kind");
  assert(node.expandable, "closure should be expandable");
  assert(node.source !== undefined, "closure should have source");
  assert(node.children !== undefined, "closure should have children from captures");
  assertEqual(node.children!.length, 1, "closure should have 1 capture");
  assertEqual(node.children![0].name, "value", "capture child name");
  assert(node.children![0].expandable, "capture with RefValue should be expandable");
  assertEqual(node.children![0].objectId, "abc:13", "capture objectId");

  console.log("  testDecodeClosureWithCaptures: PASS");
}

function testDecodeFuncNoCapturesNotClosure(): void {
  // Regular func with no captures — should be "func" not "closure"
  const node = decodeTypedValue("init", {
    T: { "@type": "/gno.FuncType", Params: [], Results: [] },
    V: {
      "@type": "/gno.FuncValue",
      Type: { "@type": "/gno.FuncType", Params: [], Results: [] },
      Name: "init",
      Captures: [],
      Source: { "@type": "/gno.RefNode", Location: { PkgPath: "gno.land/r/test", File: "test.gno", Span: { Pos: { Line: "1", Column: "1" }, End: { Line: "3", Column: "1" }, Num: "0" } }, BlockNode: null },
    },
  });

  assertEqual(node.kind, "func", "func without captures is not closure");
  assert(node.children === undefined || node.children.length === 0, "no children for non-closure");

  console.log("  testDecodeFuncNoCapturesNotClosure: PASS");
}

function testDecodeClosureMultipleCaptures(): void {
  // Closure with multiple captures
  const node = decodeTypedValue("accumulator", {
    T: { "@type": "/gno.FuncType", Params: [{ Name: "val", Type: { "@type": "/gno.PrimitiveType", value: "32" }, Embedded: false, Tag: "" }], Results: [] },
    V: {
      "@type": "/gno.FuncValue",
      Type: { "@type": "/gno.FuncType", Params: [{ Name: "val", Type: { "@type": "/gno.PrimitiveType", value: "32" }, Embedded: false, Tag: "" }], Results: [] },
      Name: "",
      Captures: [
        { T: { "@type": "/gno.heapItemType" }, V: { "@type": "/gno.RefValue", ObjectID: "abc:16", Hash: "aaa" } },
        { T: { "@type": "/gno.heapItemType" }, V: { "@type": "/gno.RefValue", ObjectID: "abc:17", Hash: "bbb" } },
      ],
      Source: { "@type": "/gno.RefNode", Location: { PkgPath: "gno.land/r/test", File: "test.gno", Span: { Pos: { Line: "23", Column: "16" }, End: { Line: "25", Column: "3" }, Num: "0" } }, BlockNode: null },
    },
  });

  assertEqual(node.kind, "closure", "multi-capture closure kind");
  assertEqual(node.children!.length, 2, "should have 2 captures");
  assertEqual(node.children![0].objectId, "abc:16", "first capture objectId");
  assertEqual(node.children![1].objectId, "abc:17", "second capture objectId");

  console.log("  testDecodeClosureMultipleCaptures: PASS");
}

// ---- Run all tests ----

console.log("decode.test.ts:");
testDecodePkg();
testDecodeObject();
testDecodeHeapItem();
testDecodeMap();
testDecodePointerRefValue();
testDecodeInlineStruct();
testDecodeCycleRef();
testDecodeNil();
testDecodeSliceRefBase();
testDecodeFuncInline();
testDecodeFuncRefValue();
testDecodeClosureWithCaptures();
testDecodeFuncNoCapturesNotClosure();
testDecodeClosureMultipleCaptures();
console.log("All tests passed.");
