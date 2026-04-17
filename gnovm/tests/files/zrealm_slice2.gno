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
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:7|blk.bs.heap](5)=
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
