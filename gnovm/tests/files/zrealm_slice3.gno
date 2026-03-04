// PKGPATH: gno.land/r/test
package test

// Verify that append to a realm []byte (Data-backed, within capacity) marks the
// backing array dirty. The DidUpdate was moved into the Data branch; this confirms
// it still fires there (raw byte copy bypasses Assign2).
var bs []byte

func init() {
	bs = make([]byte, 2, 5)
	bs[0] = 10
	bs[1] = 20
}

func main(cur realm) {
	bs = append(bs, 30)
	println("len:", len(bs))
	for i, b := range bs {
		println(i, b)
	}
}

// Output:
// len: 3
// 0 10
// 1 20
// 2 30

// Realm:
// finalizerealm["gno.land/r/test"]
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:7](5)=
//     @@ -1,10 +1,10 @@
//      {
//     -    "Data": "ChQAAAA=",
//     +    "Data": "ChQeAAA=",
//          "List": null,
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7",
//              "LastObjectSize": "182",
//     -        "ModTime": "0",
//     +        "ModTime": "7",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3",
//              "RefCount": "1"
//          }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:3](0)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3",
//              "LastObjectSize": "374",
//     -        "ModTime": "6",
//     +        "ModTime": "7",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "RefCount": "1"
//          },
//     @@ -19,10 +19,10 @@
//                  "@type": "/gno.SliceValue",
//                  "Base": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "20ce1f21d3d663cac2590f91d38e5d900c6efd25",
//     +                "Hash": "6912ee8ba7233d9d5653d4ecc89e00eadf330e34",
//                      "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7"
//                  },
//     -            "Length": "2",
//     +            "Length": "3",
//                  "Maxcap": "5",
//                  "Offset": "0"
//              }
