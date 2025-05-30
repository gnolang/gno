// PKGPATH: gno.land/r/map_test
package map_test

type Bar struct {
	name string
}

type Foo struct {
	b   Bar
	num int
}

var m = map[Foo]int{}

func main() {
	// f is neal real
	f := Foo{
		b: Bar{
			name: "a",
		},
		num: 1,
	}
	m[f] = 1
	println(m)
}

// Output:
// map{(struct{(struct{("a" string)} gno.land/r/map_test.Bar),(1 int)} gno.land/r/map_test.Foo):(1 int)}

// Realm:
// finalizerealm["gno.land/r/map_test"]
// c[02f0842ff638606e376be560152a77adbd377f56:8]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "a"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "02f0842ff638606e376be560152a77adbd377f56:8",
//         "ModTime": "0",
//         "OwnerID": "02f0842ff638606e376be560152a77adbd377f56:7",
//         "RefCount": "1"
//     }
// }
// c[02f0842ff638606e376be560152a77adbd377f56:7]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/map_test.Bar"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "a4e915a3f8a66544fda53ad59e464d0e2d8615ea",
//                 "ObjectID": "02f0842ff638606e376be560152a77adbd377f56:8"
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
//         "ID": "02f0842ff638606e376be560152a77adbd377f56:7",
//         "ModTime": "0",
//         "OwnerID": "02f0842ff638606e376be560152a77adbd377f56:4",
//         "RefCount": "1"
//     }
// }
// u[02f0842ff638606e376be560152a77adbd377f56:4]=
//     @@ -1,10 +1,31 @@
//      {
//          "List": {
//     -        "List": []
//     +        "List": [
//     +            {
//     +                "Key": {
//     +                    "T": {
//     +                        "@type": "/gno.RefType",
//     +                        "ID": "gno.land/r/map_test.Foo"
//     +                    },
//     +                    "V": {
//     +                        "@type": "/gno.RefValue",
//     +                        "Hash": "e38d41c819dfc57e9fce8018a38ea67f31813a17",
//     +                        "ObjectID": "02f0842ff638606e376be560152a77adbd377f56:7"
//     +                    }
//     +                },
//     +                "Value": {
//     +                    "N": "AQAAAAAAAAA=",
//     +                    "T": {
//     +                        "@type": "/gno.PrimitiveType",
//     +                        "value": "32"
//     +                    }
//     +                }
//     +            }
//     +        ]
//          },
//          "ObjectInfo": {
//              "ID": "02f0842ff638606e376be560152a77adbd377f56:4",
//     -        "ModTime": "0",
//     +        "ModTime": "6",
//              "OwnerID": "02f0842ff638606e376be560152a77adbd377f56:3",
//              "RefCount": "1"
//          }
