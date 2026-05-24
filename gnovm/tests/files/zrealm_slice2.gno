// PKGPATH: gno.land/r/test
package test

// Verify that copy into a realm []byte (Data-backed, via make) marks the backing
// array dirty. Assign2 skips DidUpdate for DataByteType, so the top-level
// DidUpdate(dstBase, nil, nil) in copy is the only thing ensuring persistence.
var bs []byte

func init() {
	bs = make([]byte, 5)
	bs[0] = 1
	bs[1] = 2
	bs[2] = 3
	bs[3] = 4
	bs[4] = 5
}

func main(cur realm) {
	n := copy(bs, []byte{99, 88, 77})
	println("copied:", n)
	for i, b := range bs {
		println(i, b)
	}
}

// Output:
// copied: 3
// 0 99
// 1 88
// 2 77
// 3 4
// 4 5

// Realm:
// finalizerealm["gno.land/r/test"]
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:7](5)=
//     @@ -1,10 +1,10 @@
//      {
//     -    "Data": "AQIDBAU=",
//     +    "Data": "Y1hNBAU=",
//          "List": null,
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:7",
//              "LastObjectSize": "182",
//     -        "ModTime": "0",
//     +        "ModTime": "7",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:3",
//              "RefCount": "1"
//          }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:3](0)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:3",
//              "LastObjectSize": "374",
//     -        "ModTime": "6",
//     +        "ModTime": "7",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "RefCount": "1"
//          },
//     @@ -19,7 +19,7 @@
//                  "@type": "/gno.SliceValue",
//                  "Base": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "94663b644030bc570a98dec63ae9699057c70841",
//     +                "Hash": "7b567dca03873f7b8707538e68b5f2fd0abf4b9d",
//                      "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:7"
//                  },
//                  "Length": "5",
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:2](0)=
//     @@ -34,7 +34,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "c556901dbb96baf34c4cde338551ca8af5ee4a83",
//     +                "Hash": "820bb75daa56181bdef777ad2646281f236ee761",
//                      "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:3"
//                  }
//              },
