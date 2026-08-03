// PKGPATH: gno.land/r/test
package test

// Verify that plain byte-index assignment (bs[i] = v) on a Data-backed []byte
// realm variable marks the backing ArrayValue dirty via DidUpdate, so the
// mutation persists across transactions.
var bs = []byte("aabb")

func main(cur realm) {
	bs[0] = 'A'
	bs[1] = 'B'
	println(string(bs))
}

// Output:
// ABbb

// Realm:
// finalizerealm["gno.land/r/test"]
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:4](5)=
//     @@ -1,10 +1,10 @@
//      {
//     -    "Data": "YWFiYg==",
//     +    "Data": "QUJiYg==",
//          "List": null,
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:4",
//              "LastObjectSize": "181",
//     -        "ModTime": "0",
//     +        "ModTime": "6",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:3",
//              "RefCount": "1"
//          }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:3](5)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:3",
//              "LastObjectSize": "369",
//     -        "ModTime": "0",
//     +        "ModTime": "6",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "RefCount": "1"
//          },
//     @@ -19,7 +19,7 @@
//                  "@type": "/gno.SliceValue",
//                  "Base": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "802b0b538699aeedf3f042de85c84ecd41306767",
//     +                "Hash": "ec7241d8fb1a204bbedd919540eee0999fdd47f8",
//                      "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:4"
//                  },
//                  "Length": "4",
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:2](5)=
//     @@ -4,7 +4,7 @@
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "IsEscaped": true,
//              "LastObjectSize": "537",
//     -        "ModTime": "0",
//     +        "ModTime": "6",
//              "RefCount": "2"
//          },
//          "Parent": null,
//     @@ -34,7 +34,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "48ad1a6cda3adcab93691fa4f02adc6bf248c8ad",
//     +                "Hash": "0257302f416d8ba0d121689b3b0251ac388135d4",
//                      "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:3"
//                  }
//              },
