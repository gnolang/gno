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
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:7](5)=
//     @@ -1,10 +1,10 @@
//      {
//     -    "Data": "AQIDBAU=",
//     +    "Data": "Y1hNBAU=",
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
//     @@ -19,7 +19,7 @@
//                  "@type": "/gno.SliceValue",
//                  "Base": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "6010b302f15af22d9b01c018881b3c89469c094a",
//     +                "Hash": "4aa1e1fd4cb1626283e04b783e840ee68ed5d49b",
//                      "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7"
//                  },
//                  "Length": "5",
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:2](0)=
//     @@ -34,7 +34,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "c0a44164f7316678b2ed0f68b3fc9865553c4790",
//     +                "Hash": "d44b07317bf9df7bc963c082ff2ff159fe5d1d6c",
//                      "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3"
//                  }
//              },
