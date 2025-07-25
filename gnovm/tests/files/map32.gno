// PKGPATH: gno.land/r/mapkey
package mapkey

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
// map{(struct{(struct{("a" string)} gno.land/r/mapkey.Bar),(1 int)} gno.land/r/mapkey.Foo):(1 int)}

// Realm:
// finalizerealm["gno.land/r/mapkey"]
// c[2a17fa0a22a6e119cfaf6e864b74063e0e4d464d:8:0](243)={
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
//         "ID": "2a17fa0a22a6e119cfaf6e864b74063e0e4d464d:8:0",
//         "LastObjectSize": "243",
//         "ModTime": "0",
//         "OwnerID": "2a17fa0a22a6e119cfaf6e864b74063e0e4d464d:7:0",
//         "RefCount": "1"
//     }
// }
// c[2a17fa0a22a6e119cfaf6e864b74063e0e4d464d:7:0](379)={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/mapkey.Bar"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "d5bc5ce52ab96d39523f44ca361d489f4cb8a983",
//                 "ObjectID": "2a17fa0a22a6e119cfaf6e864b74063e0e4d464d:8:0"
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
//         "ID": "2a17fa0a22a6e119cfaf6e864b74063e0e4d464d:7:0",
//         "LastObjectSize": "379",
//         "ModTime": "0",
//         "OwnerID": "2a17fa0a22a6e119cfaf6e864b74063e0e4d464d:4:0",
//         "RefCount": "1"
//     }
// }
// u[2a17fa0a22a6e119cfaf6e864b74063e0e4d464d:4:0](208)=
//     @@ -1,11 +1,32 @@
//      {
//          "List": {
//     -        "List": []
//     +        "List": [
//     +            {
//     +                "Key": {
//     +                    "T": {
//     +                        "@type": "/gno.RefType",
//     +                        "ID": "gno.land/r/mapkey.Foo"
//     +                    },
//     +                    "V": {
//     +                        "@type": "/gno.RefValue",
//     +                        "Hash": "97fd16bb401f556b356dbdf4d9ab72a3a74efa02",
//     +                        "ObjectID": "2a17fa0a22a6e119cfaf6e864b74063e0e4d464d:7:0"
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
//              "ID": "2a17fa0a22a6e119cfaf6e864b74063e0e4d464d:4:0",
//              "LastObjectSize": "179",
//     -        "ModTime": "0",
//     +        "ModTime": "6",
//              "OwnerID": "2a17fa0a22a6e119cfaf6e864b74063e0e4d464d:3:0",
//              "RefCount": "1"
//          }
