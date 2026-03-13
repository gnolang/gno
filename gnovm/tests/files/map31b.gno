// PKGPATH: gno.land/r/mapkey
package mapkey

var (
	m = map[*int]int{}
	i = 3 // real
)

func main() {
	tmpInt := &i
	m[tmpInt] = 1
}

// Realm:
// finalizerealm["gno.land/r/mapkey"]
// u[2a17fa0a22a6e119cfaf6e864b74063e0e4d464d:4](239)=
//     @@ -1,11 +1,40 @@
//      {
//          "List": {
//     -        "List": []
//     +        "List": [
//     +            {
//     +                "Key": {
//     +                    "T": {
//     +                        "@type": "/gno.PointerType",
//     +                        "Elt": {
//     +                            "@type": "/gno.PrimitiveType",
//     +                            "value": "32"
//     +                        }
//     +                    },
//     +                    "V": {
//     +                        "@type": "/gno.PointerValue",
//     +                        "Base": {
//     +                            "@type": "/gno.RefValue",
//     +                            "Escaped": true,
//     +                            "ObjectID": "2a17fa0a22a6e119cfaf6e864b74063e0e4d464d:5"
//     +                        },
//     +                        "Index": "0",
//     +                        "TV": null
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
//     +        "ModTime": "7",
//              "OwnerID": "2a17fa0a22a6e119cfaf6e864b74063e0e4d464d:3",
//              "RefCount": "1"
//          }
// u[2a17fa0a22a6e119cfaf6e864b74063e0e4d464d:5](7)=
//     @@ -1,10 +1,10 @@
//      {
//          "ObjectInfo": {
//              "ID": "2a17fa0a22a6e119cfaf6e864b74063e0e4d464d:5",
//     +        "IsEscaped": true,
//              "LastObjectSize": "216",
//     -        "ModTime": "0",
//     -        "OwnerID": "2a17fa0a22a6e119cfaf6e864b74063e0e4d464d:2",
//     -        "RefCount": "1"
//     +        "ModTime": "7",
//     +        "RefCount": "2"
//          },
//          "Value": {
//              "N": "AwAAAAAAAAA=",
// u[2a17fa0a22a6e119cfaf6e864b74063e0e4d464d:2](7)=
//     @@ -4,7 +4,7 @@
//              "ID": "2a17fa0a22a6e119cfaf6e864b74063e0e4d464d:2",
//              "IsEscaped": true,
//              "LastObjectSize": "636",
//     -        "ModTime": "0",
//     +        "ModTime": "7",
//              "RefCount": "2"
//          },
//          "Parent": null,
//     @@ -34,7 +34,6 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "d9bfc05ece20a632345b370988b3fb6a5afd7651",
//                      "ObjectID": "2a17fa0a22a6e119cfaf6e864b74063e0e4d464d:3"
//                  }
//              },
//     @@ -44,7 +43,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "5674f063ba1ebd8e48e658855dbf850a00c118c7",
//     +                "Escaped": true,
//                      "ObjectID": "2a17fa0a22a6e119cfaf6e864b74063e0e4d464d:5"
//                  }
//              },
// u[2a17fa0a22a6e119cfaf6e864b74063e0e4d464d:3](5)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "2a17fa0a22a6e119cfaf6e864b74063e0e4d464d:3",
//              "LastObjectSize": "388",
//     -        "ModTime": "0",
//     +        "ModTime": "7",
//              "OwnerID": "2a17fa0a22a6e119cfaf6e864b74063e0e4d464d:2",
//              "RefCount": "1"
//          },
//     @@ -23,7 +23,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "57424a2a4c05dd2e2f21e0c76ab122ccd15830ce",
//     +            "Hash": "c9b0662ab7453d47ca8cac636d4812295aad9e1f",
//                  "ObjectID": "2a17fa0a22a6e119cfaf6e864b74063e0e4d464d:4"
//              }
//          }
