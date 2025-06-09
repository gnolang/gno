package main

import "gno.land/r/demo/emit"

func main() {
	emit.Emit("foo")
	emit.Emit("bar")
}

// Events:
// [
//   {
//     "type": "EventName",
//     "attrs": [
//       {
//         "key": "key",
//         "value": "foo"
//       }
//     ],
//     "pkg_path": "gno.land/r/demo/emit"
//   },
//   {
//     "type": "EventName",
//     "attrs": [
//       {
//         "key": "key",
//         "value": "bar"
//       }
//     ],
//     "pkg_path": "gno.land/r/demo/emit"
//   }
// ]
