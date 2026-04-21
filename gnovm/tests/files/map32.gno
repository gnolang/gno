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
// c[0a17fa0a22a6e119cfaf6e864b74063e0e4d464d:8](239)={
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
//         "ID": "0a17fa0a22a6e119cfaf6e864b74063e0e4d464d:8",
//         "LastObjectSize": "239",
//         "ModTime": "0",
//         "OwnerID": "0a17fa0a22a6e119cfaf6e864b74063e0e4d464d:7",
//         "RefCount": "1"
//     }
// }
// c[0a17fa0a22a6e119cfaf6e864b74063e0e4d464d:7](373)={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/mapkey.Bar"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "e7c6b15cf175010b3d14d7a18e68808b5144bb39",
//                 "ObjectID": "0a17fa0a22a6e119cfaf6e864b74063e0e4d464d:8"
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
//         "ID": "0a17fa0a22a6e119cfaf6e864b74063e0e4d464d:7",
//         "LastObjectSize": "373",
//         "ModTime": "0",
//         "OwnerID": "0a17fa0a22a6e119cfaf6e864b74063e0e4d464d:4",
//         "RefCount": "1"
//     }
// }
// u[0a17fa0a22a6e119cfaf6e864b74063e0e4d464d:4](206)=
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
//     +                        "Hash": "019dd1c84943ce603928bf72d02a9c3fd785501d",
//     +                        "ObjectID": "0a17fa0a22a6e119cfaf6e864b74063e0e4d464d:7"
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
//              "ID": "0a17fa0a22a6e119cfaf6e864b74063e0e4d464d:4",
//              "LastObjectSize": "175",
//     -        "ModTime": "0",
//     +        "ModTime": "6",
//              "OwnerID": "0a17fa0a22a6e119cfaf6e864b74063e0e4d464d:3",
//              "RefCount": "1"
//          }
