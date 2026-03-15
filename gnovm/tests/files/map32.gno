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
// c[2a17fa0a22a6e119cfaf6e864b74063e0e4d464d:8](239)={
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
//         "ID": "2a17fa0a22a6e119cfaf6e864b74063e0e4d464d:8",
//         "LastObjectSize": "239",
//         "ModTime": "0",
//         "OwnerID": "2a17fa0a22a6e119cfaf6e864b74063e0e4d464d:7",
//         "RefCount": "1"
//     }
// }
// c[2a17fa0a22a6e119cfaf6e864b74063e0e4d464d:7](373)={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/mapkey.Bar"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "7bfb3c98655cc8320fb2b037dc18cdb19a2a184e",
//                 "ObjectID": "2a17fa0a22a6e119cfaf6e864b74063e0e4d464d:8"
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
//         "ID": "2a17fa0a22a6e119cfaf6e864b74063e0e4d464d:7",
//         "LastObjectSize": "373",
//         "ModTime": "0",
//         "OwnerID": "2a17fa0a22a6e119cfaf6e864b74063e0e4d464d:4",
//         "RefCount": "1"
//     }
// }
// u[2a17fa0a22a6e119cfaf6e864b74063e0e4d464d:4](206)=
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
//     +                        "Hash": "faf067a62585266c9bfac0c93d2d03573c24fceb",
//     +                        "ObjectID": "2a17fa0a22a6e119cfaf6e864b74063e0e4d464d:7"
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
//              "ID": "2a17fa0a22a6e119cfaf6e864b74063e0e4d464d:4",
//              "LastObjectSize": "175",
//     -        "ModTime": "0",
//     +        "ModTime": "6",
//              "OwnerID": "2a17fa0a22a6e119cfaf6e864b74063e0e4d464d:3",
//              "RefCount": "1"
//          }
// u[2a17fa0a22a6e119cfaf6e864b74063e0e4d464d:3](5)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "2a17fa0a22a6e119cfaf6e864b74063e0e4d464d:3",
//              "LastObjectSize": "381",
//     -        "ModTime": "0",
//     +        "ModTime": "8",
//              "OwnerID": "2a17fa0a22a6e119cfaf6e864b74063e0e4d464d:2",
//              "RefCount": "1"
//          },
//     @@ -20,7 +20,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "57424a2a4c05dd2e2f21e0c76ab122ccd15830ce",
//     +            "Hash": "fe324dbfb01c891c641a7d1f1a672b0374e0eadf",
//                  "ObjectID": "2a17fa0a22a6e119cfaf6e864b74063e0e4d464d:4"
//              }
//          }
// u[2a17fa0a22a6e119cfaf6e864b74063e0e4d464d:2](5)=
//     @@ -4,7 +4,7 @@
//              "ID": "2a17fa0a22a6e119cfaf6e864b74063e0e4d464d:2",
//              "IsEscaped": true,
//              "LastObjectSize": "889",
//     -        "ModTime": "0",
//     +        "ModTime": "8",
//              "RefCount": "2"
//          },
//          "Parent": null,
//     @@ -131,7 +131,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "07e7084073c41506278b9c96c7064310b043b963",
//     +                "Hash": "c63a45a7727378197b1d42e85c28a782cb9c7fdd",
//                      "ObjectID": "2a17fa0a22a6e119cfaf6e864b74063e0e4d464d:3"
//                  }
//              },
