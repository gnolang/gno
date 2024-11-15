package main

func main() {
	event.Emit("foo")
	event.Emit("bar")
}

// Events:
// [
//   {
//     "type": "TAG",
//     "attrs": [
//       {
//         "key": "key",
//         "value": "foo"
//       }
//     ],
//     "pkg_path": "gno.land/r/demo/event",
//     "func": "Emit"
//   },
//   {
//     "type": "TAG",
//     "attrs": [
//       {
//         "key": "key",
//         "value": "bar"
//       }
//     ],
//     "pkg_path": "gno.land/r/demo/event",
//     "func": "Emit"
//   }
// ]
