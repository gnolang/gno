// PKGPATH: gno.land/r/test
package test

type T struct{ x int }

func (t T) Get() int { return t.x }

var (
	M1  any // map with local-type key
	M2  any // map with local-type value
	SL  any // slice of local type
	P   any // pointer to local type
	RN  any // recursive local type
	NB  any // nested local types
	MI  any // primitive-based local type
	FT  any // local func type
	G   func() int
	ARR any // byte-based local type array
)

func main(cur realm) {
	// 1 map key
	type K struct{ k int }
	m := map[K]string{{1}: "one"}
	M1 = m
	println(m[K{1}])
	// 2 map value
	type V struct{ v int }
	mv := map[string]V{"a": {2}}
	M2 = mv
	println(mv["a"].v)
	// 3 slice of local type
	type S struct{ x int }
	sl := []S{{1}, {2}}
	SL = sl
	println(sl[0].x + sl[1].x)
	// 4 pointer to local type
	p := &S{5}
	P = p
	println(p.x)
	// 5 recursive local type
	type N struct {
		next *N
		v    int
	}
	n := &N{&N{nil, 2}, 1}
	RN = n
	sum := 0
	for ; n != nil; n = n.next {
		sum += n.v
	}
	println(sum)
	// 6 nested local types
	type A struct{ v int }
	type B struct{ a A }
	b := B{A{6}}
	NB = b
	println(b.a.v)
	// 7 primitive base
	type MyInt int
	MI = MyInt(42)
	println(int(MI.(MyInt)))
	// 8 local func type
	type F func() int
	var f F = func() int { return 8 }
	FT = f
	println(f())
	// 9 closure over var of local interface type
	type LI interface{ Get() int }
	var li LI = T{9}
	G = func() int { return li.Get() }
	println(G())
	// 10 byte-based local type array
	type MB byte
	arr := [3]MB{1, 2, 3}
	ARR = arr
	println(int(arr[0]) + int(arr[1]) + int(arr[2]))
}

// Output:
// one
// 2
// 3
// 5
// 3
// 6
// 42
// 8
// 9
// 6

// Realm:
// finalizerealm["gno.land/r/test"]
// c[08ada09dee16d791fd406d629fe29bb0ed084a30:16](216)={
//     "Fields": [
//         {
//             "N": "AQAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "32"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:16",
//         "LastObjectSize": "216",
//         "ModTime": "0",
//         "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:15",
//         "RefCount": "1"
//     }
// }
// c[08ada09dee16d791fd406d629fe29bb0ed084a30:15](450)={
//     "List": {
//         "List": [
//             {
//                 "Key": {
//                     "T": {
//                         "@type": "/gno.RefType",
//                         "ID": "gno.land/r/test[gno.land/r/test/zrealm_localtype1.gno:21:1-78:2].K"
//                     },
//                     "V": {
//                         "@type": "/gno.RefValue",
//                         "Hash": "3a31095f600f8196cd0da9b43120e871cf3b495d",
//                         "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:16"
//                     }
//                 },
//                 "Value": {
//                     "T": {
//                         "@type": "/gno.PrimitiveType",
//                         "value": "16"
//                     },
//                     "V": {
//                         "@type": "/gno.StringValue",
//                         "value": "one"
//                     }
//                 }
//             }
//         ]
//     },
//     "ObjectInfo": {
//         "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:15",
//         "LastObjectSize": "450",
//         "ModTime": "0",
//         "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:3",
//         "RefCount": "1"
//     }
// }
// c[08ada09dee16d791fd406d629fe29bb0ed084a30:18](216)={
//     "Fields": [
//         {
//             "N": "AgAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "32"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:18",
//         "LastObjectSize": "216",
//         "ModTime": "0",
//         "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:17",
//         "RefCount": "1"
//     }
// }
// c[08ada09dee16d791fd406d629fe29bb0ed084a30:17](448)={
//     "List": {
//         "List": [
//             {
//                 "Key": {
//                     "T": {
//                         "@type": "/gno.PrimitiveType",
//                         "value": "16"
//                     },
//                     "V": {
//                         "@type": "/gno.StringValue",
//                         "value": "a"
//                     }
//                 },
//                 "Value": {
//                     "T": {
//                         "@type": "/gno.RefType",
//                         "ID": "gno.land/r/test[gno.land/r/test/zrealm_localtype1.gno:21:1-78:2].V"
//                     },
//                     "V": {
//                         "@type": "/gno.RefValue",
//                         "Hash": "77666b02a600218029ec00491b3b24d35da21152",
//                         "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:18"
//                     }
//                 }
//             }
//         ]
//     },
//     "ObjectInfo": {
//         "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:17",
//         "LastObjectSize": "448",
//         "ModTime": "0",
//         "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:4",
//         "RefCount": "1"
//     }
// }
// c[08ada09dee16d791fd406d629fe29bb0ed084a30:20](216)={
//     "Fields": [
//         {
//             "N": "AQAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "32"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:20",
//         "LastObjectSize": "216",
//         "ModTime": "0",
//         "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:19",
//         "RefCount": "1"
//     }
// }
// c[08ada09dee16d791fd406d629fe29bb0ed084a30:21](216)={
//     "Fields": [
//         {
//             "N": "AgAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "32"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:21",
//         "LastObjectSize": "216",
//         "ModTime": "0",
//         "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:19",
//         "RefCount": "1"
//     }
// }
// c[08ada09dee16d791fd406d629fe29bb0ed084a30:19](586)={
//     "Data": null,
//     "List": [
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/test[gno.land/r/test/zrealm_localtype1.gno:21:1-78:2].S"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "d103e88bdf4a480a139215347a42fff8b518f96f",
//                 "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:20"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/test[gno.land/r/test/zrealm_localtype1.gno:21:1-78:2].S"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "ba68cc4c259baf6c87ce54d7cd71d1b94c4e081c",
//                 "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:21"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:19",
//         "LastObjectSize": "586",
//         "ModTime": "0",
//         "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:5",
//         "RefCount": "1"
//     }
// }
// c[08ada09dee16d791fd406d629fe29bb0ed084a30:23](216)={
//     "Fields": [
//         {
//             "N": "BQAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "32"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:23",
//         "LastObjectSize": "216",
//         "ModTime": "0",
//         "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:22",
//         "RefCount": "1"
//     }
// }
// c[08ada09dee16d791fd406d629fe29bb0ed084a30:22](384)={
//     "ObjectInfo": {
//         "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:22",
//         "LastObjectSize": "384",
//         "ModTime": "0",
//         "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:6",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/test[gno.land/r/test/zrealm_localtype1.gno:21:1-78:2].S"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "9715ebddfea624058a9cb5f18831dc66405437ae",
//             "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:23"
//         }
//     }
// }
// c[08ada09dee16d791fd406d629fe29bb0ed084a30:27](336)={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/test[gno.land/r/test/zrealm_localtype1.gno:21:1-78:2].N"
//                 }
//             }
//         },
//         {
//             "N": "AgAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "32"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:27",
//         "LastObjectSize": "336",
//         "ModTime": "0",
//         "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:26",
//         "RefCount": "1"
//     }
// }
// c[08ada09dee16d791fd406d629fe29bb0ed084a30:26](385)={
//     "ObjectInfo": {
//         "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:26",
//         "LastObjectSize": "385",
//         "ModTime": "0",
//         "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:25",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/test[gno.land/r/test/zrealm_localtype1.gno:21:1-78:2].N"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "1784f490f71fb3665d31d53459a960f58df4198a",
//             "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:27"
//         }
//     }
// }
// c[08ada09dee16d791fd406d629fe29bb0ed084a30:25](466)={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/test[gno.land/r/test/zrealm_localtype1.gno:21:1-78:2].N"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "3c4df2464f1466b9a711d4fce91f39830d86841f",
//                     "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:26"
//                 },
//                 "Index": "0",
//                 "TV": null
//             }
//         },
//         {
//             "N": "AQAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "32"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:25",
//         "LastObjectSize": "466",
//         "ModTime": "0",
//         "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:24",
//         "RefCount": "1"
//     }
// }
// c[08ada09dee16d791fd406d629fe29bb0ed084a30:24](384)={
//     "ObjectInfo": {
//         "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:24",
//         "LastObjectSize": "384",
//         "ModTime": "0",
//         "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:7",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/test[gno.land/r/test/zrealm_localtype1.gno:21:1-78:2].N"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "9707898483a5b4ad5ebe5877e7320b053c482197",
//             "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:25"
//         }
//     }
// }
// c[08ada09dee16d791fd406d629fe29bb0ed084a30:29](216)={
//     "Fields": [
//         {
//             "N": "BgAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "32"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:29",
//         "LastObjectSize": "216",
//         "ModTime": "0",
//         "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:28",
//         "RefCount": "1"
//     }
// }
// c[08ada09dee16d791fd406d629fe29bb0ed084a30:28](382)={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/test[gno.land/r/test/zrealm_localtype1.gno:21:1-78:2].A"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "381d92a21a9cbb6f19706c9feeda0a19db4c73b3",
//                 "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:29"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:28",
//         "LastObjectSize": "382",
//         "ModTime": "0",
//         "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:8",
//         "RefCount": "1"
//     }
// }
// c[08ada09dee16d791fd406d629fe29bb0ed084a30:30](326)={
//     "Crossing": false,
//     "FileName": "",
//     "IsClosure": true,
//     "IsMethod": false,
//     "Name": "",
//     "NativeName": "",
//     "NativePkg": "",
//     "ObjectInfo": {
//         "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:30",
//         "LastObjectSize": "326",
//         "ModTime": "0",
//         "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:10",
//         "RefCount": "1"
//     },
//     "Parent": null,
//     "PkgPath": "gno.land/r/test",
//     "Source": {
//         "@type": "/gno.RefNode",
//         "BlockNode": null,
//         "Location": {
//             "File": "zrealm_localtype1.gno",
//             "PkgPath": "gno.land/r/test",
//             "Span": {
//                 "End": {
//                     "Column": "35",
//                     "Line": "65"
//                 },
//                 "Num": "0",
//                 "Pos": {
//                     "Column": "12",
//                     "Line": "65"
//                 }
//             }
//         }
//     },
//     "Type": {
//         "@type": "/gno.FuncType",
//         "Params": [],
//         "Results": [
//             {
//                 "Embedded": false,
//                 "Name": ".res.0",
//                 "PkgPath": "",
//                 "Tag": "",
//                 "Type": {
//                     "@type": "/gno.PrimitiveType",
//                     "value": "32"
//                 }
//             }
//         ]
//     }
// }
// c[08ada09dee16d791fd406d629fe29bb0ed084a30:33](216)={
//     "Fields": [
//         {
//             "N": "CQAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "32"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:33",
//         "LastObjectSize": "216",
//         "ModTime": "0",
//         "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:32",
//         "RefCount": "1"
//     }
// }
// c[08ada09dee16d791fd406d629fe29bb0ed084a30:32](336)={
//     "ObjectInfo": {
//         "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:32",
//         "LastObjectSize": "336",
//         "ModTime": "0",
//         "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:31",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/test.T"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "a690d637ed6d47cdec5f4b326480c1ac05b93f48",
//             "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:33"
//         }
//     }
// }
// c[08ada09dee16d791fd406d629fe29bb0ed084a30:31](466)={
//     "Captures": [
//         {
//             "T": {
//                 "@type": "/gno.heapItemType"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "03a5a4953722aa3012cf75ff98a6fd7f5baf5b30",
//                 "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:32"
//             }
//         }
//     ],
//     "Crossing": false,
//     "FileName": "",
//     "IsClosure": true,
//     "IsMethod": false,
//     "Name": "",
//     "NativeName": "",
//     "NativePkg": "",
//     "ObjectInfo": {
//         "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:31",
//         "LastObjectSize": "466",
//         "ModTime": "0",
//         "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:11",
//         "RefCount": "1"
//     },
//     "Parent": null,
//     "PkgPath": "gno.land/r/test",
//     "Source": {
//         "@type": "/gno.RefNode",
//         "BlockNode": null,
//         "Location": {
//             "File": "zrealm_localtype1.gno",
//             "PkgPath": "gno.land/r/test",
//             "Span": {
//                 "End": {
//                     "Column": "36",
//                     "Line": "71"
//                 },
//                 "Num": "0",
//                 "Pos": {
//                     "Column": "6",
//                     "Line": "71"
//                 }
//             }
//         }
//     },
//     "Type": {
//         "@type": "/gno.FuncType",
//         "Params": [],
//         "Results": [
//             {
//                 "Embedded": false,
//                 "Name": ".res.0",
//                 "PkgPath": "",
//                 "Tag": "",
//                 "Type": {
//                     "@type": "/gno.PrimitiveType",
//                     "value": "32"
//                 }
//             }
//         ]
//     }
// }
// c[08ada09dee16d791fd406d629fe29bb0ed084a30:34](182)={
//     "Data": "AQID",
//     "List": null,
//     "ObjectInfo": {
//         "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:34",
//         "LastObjectSize": "182",
//         "ModTime": "0",
//         "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:12",
//         "RefCount": "1"
//     }
// }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:3](243)=
//     @@ -2,9 +2,26 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:3",
//              "LastObjectSize": "190",
//     -        "ModTime": "0",
//     +        "ModTime": "14",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "RefCount": "1"
//          },
//     -    "Value": {}
//     +    "Value": {
//     +        "T": {
//     +            "@type": "/gno.MapType",
//     +            "Key": {
//     +                "@type": "/gno.RefType",
//     +                "ID": "gno.land/r/test[gno.land/r/test/zrealm_localtype1.gno:21:1-78:2].K"
//     +            },
//     +            "Value": {
//     +                "@type": "/gno.PrimitiveType",
//     +                "value": "16"
//     +            }
//     +        },
//     +        "V": {
//     +            "@type": "/gno.RefValue",
//     +            "Hash": "333805ec89565ae79c21f32596ba7934aae56bb7",
//     +            "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:15"
//     +        }
//     +    }
//      }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:4](243)=
//     @@ -2,9 +2,26 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:4",
//              "LastObjectSize": "190",
//     -        "ModTime": "0",
//     +        "ModTime": "14",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "RefCount": "1"
//          },
//     -    "Value": {}
//     +    "Value": {
//     +        "T": {
//     +            "@type": "/gno.MapType",
//     +            "Key": {
//     +                "@type": "/gno.PrimitiveType",
//     +                "value": "16"
//     +            },
//     +            "Value": {
//     +                "@type": "/gno.RefType",
//     +                "ID": "gno.land/r/test[gno.land/r/test/zrealm_localtype1.gno:21:1-78:2].V"
//     +            }
//     +        },
//     +        "V": {
//     +            "@type": "/gno.RefValue",
//     +            "Hash": "be4b1756be7827dfe1aa0829247459f04df0dbe5",
//     +            "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:17"
//     +        }
//     +    }
//      }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:5](244)=
//     @@ -2,9 +2,29 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:5",
//              "LastObjectSize": "190",
//     -        "ModTime": "0",
//     +        "ModTime": "14",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "RefCount": "1"
//          },
//     -    "Value": {}
//     +    "Value": {
//     +        "T": {
//     +            "@type": "/gno.SliceType",
//     +            "Elt": {
//     +                "@type": "/gno.RefType",
//     +                "ID": "gno.land/r/test[gno.land/r/test/zrealm_localtype1.gno:21:1-78:2].S"
//     +            },
//     +            "Vrd": false
//     +        },
//     +        "V": {
//     +            "@type": "/gno.SliceValue",
//     +            "Base": {
//     +                "@type": "/gno.RefValue",
//     +                "Hash": "092fce08dfff54c13da179a175a04d3f3cc84d66",
//     +                "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:19"
//     +            },
//     +            "Length": "2",
//     +            "Maxcap": "2",
//     +            "Offset": "0"
//     +        }
//     +    }
//      }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:6](243)=
//     @@ -2,9 +2,27 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:6",
//              "LastObjectSize": "190",
//     -        "ModTime": "0",
//     +        "ModTime": "14",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "RefCount": "1"
//          },
//     -    "Value": {}
//     +    "Value": {
//     +        "T": {
//     +            "@type": "/gno.PointerType",
//     +            "Elt": {
//     +                "@type": "/gno.RefType",
//     +                "ID": "gno.land/r/test[gno.land/r/test/zrealm_localtype1.gno:21:1-78:2].S"
//     +            }
//     +        },
//     +        "V": {
//     +            "@type": "/gno.PointerValue",
//     +            "Base": {
//     +                "@type": "/gno.RefValue",
//     +                "Hash": "fba4c501e73c06c3596d482abd01e46aef9231b5",
//     +                "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:22"
//     +            },
//     +            "Index": "0",
//     +            "TV": null
//     +        }
//     +    }
//      }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:7](243)=
//     @@ -2,9 +2,27 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:7",
//              "LastObjectSize": "190",
//     -        "ModTime": "0",
//     +        "ModTime": "14",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "RefCount": "1"
//          },
//     -    "Value": {}
//     +    "Value": {
//     +        "T": {
//     +            "@type": "/gno.PointerType",
//     +            "Elt": {
//     +                "@type": "/gno.RefType",
//     +                "ID": "gno.land/r/test[gno.land/r/test/zrealm_localtype1.gno:21:1-78:2].N"
//     +            }
//     +        },
//     +        "V": {
//     +            "@type": "/gno.PointerValue",
//     +            "Base": {
//     +                "@type": "/gno.RefValue",
//     +                "Hash": "fe48c3ee241a8f6b77437fdea17caccaf1bf0946",
//     +                "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:24"
//     +            },
//     +            "Index": "0",
//     +            "TV": null
//     +        }
//     +    }
//      }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:8](198)=
//     @@ -2,9 +2,19 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:8",
//              "LastObjectSize": "190",
//     -        "ModTime": "0",
//     +        "ModTime": "14",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "RefCount": "1"
//          },
//     -    "Value": {}
//     +    "Value": {
//     +        "T": {
//     +            "@type": "/gno.RefType",
//     +            "ID": "gno.land/r/test[gno.land/r/test/zrealm_localtype1.gno:21:1-78:2].B"
//     +        },
//     +        "V": {
//     +            "@type": "/gno.RefValue",
//     +            "Hash": "04aebe78e2c4f6f8884c868dffb3caca6d091c57",
//     +            "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:28"
//     +        }
//     +    }
//      }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:9](95)=
//     @@ -2,9 +2,15 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:9",
//              "LastObjectSize": "190",
//     -        "ModTime": "0",
//     +        "ModTime": "14",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "RefCount": "1"
//          },
//     -    "Value": {}
//     +    "Value": {
//     +        "N": "KgAAAAAAAAA=",
//     +        "T": {
//     +            "@type": "/gno.RefType",
//     +            "ID": "gno.land/r/test[gno.land/r/test/zrealm_localtype1.gno:21:1-78:2].MyInt"
//     +        }
//     +    }
//      }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:10](198)=
//     @@ -2,9 +2,19 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:10",
//              "LastObjectSize": "191",
//     -        "ModTime": "0",
//     +        "ModTime": "14",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "RefCount": "1"
//          },
//     -    "Value": {}
//     +    "Value": {
//     +        "T": {
//     +            "@type": "/gno.RefType",
//     +            "ID": "gno.land/r/test[gno.land/r/test/zrealm_localtype1.gno:21:1-78:2].F"
//     +        },
//     +        "V": {
//     +            "@type": "/gno.RefValue",
//     +            "Hash": "b20269a0301eafcddc7f33a3829b8b564762d7da",
//     +            "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:30"
//     +        }
//     +    }
//      }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:11](112)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:11",
//              "LastObjectSize": "246",
//     -        "ModTime": "0",
//     +        "ModTime": "14",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "RefCount": "1"
//          },
//     @@ -22,6 +22,11 @@
//                          }
//                      }
//                  ]
//     +        },
//     +        "V": {
//     +            "@type": "/gno.RefValue",
//     +            "Hash": "0ff933126bcb9b0721726ec7f39be19316e3b2a4",
//     +            "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:31"
//              }
//          }
//      }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:12](221)=
//     @@ -2,9 +2,24 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:12",
//              "LastObjectSize": "191",
//     -        "ModTime": "0",
//     +        "ModTime": "14",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "RefCount": "1"
//          },
//     -    "Value": {}
//     +    "Value": {
//     +        "T": {
//     +            "@type": "/gno.ArrayType",
//     +            "Elt": {
//     +                "@type": "/gno.RefType",
//     +                "ID": "gno.land/r/test[gno.land/r/test/zrealm_localtype1.gno:21:1-78:2].MB"
//     +            },
//     +            "Len": "3",
//     +            "Vrd": false
//     +        },
//     +        "V": {
//     +            "@type": "/gno.RefValue",
//     +            "Hash": "d88a4615b8c816749b44e46b1b5ce6db6c3cc2f7",
//     +            "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:34"
//     +        }
//     +    }
//      }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:2](5)=
//     @@ -4,7 +4,7 @@
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "IsEscaped": true,
//              "LastObjectSize": "1878",
//     -        "ModTime": "0",
//     +        "ModTime": "34",
//              "RefCount": "2"
//          },
//          "Parent": null,
//     @@ -46,7 +46,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "77f76390147dcb12d9de104eb549c3a959934c1c",
//     +                "Hash": "251bbbda3ca0922f021a5063d477383de1197334",
//                      "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:3"
//                  }
//              },
//     @@ -56,7 +56,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "c179e3d792fe6624f9de9aed3795cb7b597a4eac",
//     +                "Hash": "03180dc16509232e8ba5aaf03d8a49f764ea3ea9",
//                      "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:4"
//                  }
//              },
//     @@ -66,7 +66,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "091813d68ddddd3d9a460952d01c6e8427f33f07",
//     +                "Hash": "9bd2a5e95cf8830ecd9d52b035ea8a1caa10b27e",
//                      "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:5"
//                  }
//              },
//     @@ -76,7 +76,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "50a7f52f8f9d6e7936afc5575fedbf44ca9b5794",
//     +                "Hash": "07a0082a11ff4b727fc643993e4e9bdac81ffd5c",
//                      "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:6"
//                  }
//              },
//     @@ -86,7 +86,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "bd403800d0a21ad2f67e595f2c4b72fc20c3a586",
//     +                "Hash": "b3e7fcfa4fe2d40b9f2fd576c0bb8e934c31d055",
//                      "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:7"
//                  }
//              },
//     @@ -96,7 +96,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "eb9fed50ea2acd720bd703f381bf1921d249bf7e",
//     +                "Hash": "5c976faab35113697a70bcbc404b4fe73ef2c409",
//                      "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:8"
//                  }
//              },
//     @@ -106,7 +106,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "82344475eb61c0973866f39df1624309fc36d7cf",
//     +                "Hash": "cfe6625bb8d990abdf1c2a309e2db1da695e1541",
//                      "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:9"
//                  }
//              },
//     @@ -116,7 +116,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "5c457e1f499a38dc90ee998f8e08e2e03260be6e",
//     +                "Hash": "988929753d9bdc43d0df1f6dff85016beeec49d7",
//                      "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:10"
//                  }
//              },
//     @@ -126,7 +126,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "2120ffb874fea910d2be24544c0f581616844490",
//     +                "Hash": "6473c217704680c5df49fb9e5c759007a541c6da",
//                      "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:11"
//                  }
//              },
//     @@ -136,7 +136,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "ae44a091b977f7ac33c6b10feb8b832efea80956",
//     +                "Hash": "99b29000effaa15a712a147b40038f835b46d645",
//                      "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:12"
//                  }
//              },
