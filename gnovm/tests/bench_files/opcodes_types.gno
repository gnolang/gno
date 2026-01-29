package main

// Test type operations: type switch, complex types (struct with slice, map, chan, func, interface)

func main() {
	OpTypeSwitch()
	OpTypes()
}

func OpTypeSwitch() {
	values := []any{1, "hello", 1.0, any(nil)}
	for _, v := range values {
		switch v.(type) {
		case int:
			// ...
		case string:
			// ...
		case float64:
			// ...
		default:
			// ...
		}
	}
}

func OpTypes() {
	t := struct {
		a []string
		b map[string]string
		c chan string
		d func()
		e any
	}{}
	_ = t
}

// Gasprofile:
// Opcodes:
//   OpAssign: count=1 gas=79
//   OpBody: count=9 gas=387
//   OpCall: count=3 gas=768
//   OpChanType: count=2 gas=114
//   OpCompositeLit: count=2 gas=100
//   OpConvert: count=1 gas=16
//   OpDefine: count=2 gas=222
//   OpEval: count=67 gas=1943
//   OpExec: count=4 gas=100
//   OpFieldType: count=5 gas=295
//   OpFuncType: count=5 gas=405
//   OpHalt: count=35 gas=35
//   OpMapType: count=2 gas=118
//   OpPopBlock: count=38 gas=114
//   OpPopFrameAndReset: count=4 gas=60
//   OpPopResults: count=3 gas=3
//   OpPrecall: count=4 gas=828
//   OpRangeIter: count=8 gas=840
//   OpReturnFromBlock: count=3 gas=108
//   OpSliceLit: count=1 gas=183
//   OpSliceType: count=3 gas=165
//   OpStructLit: count=1 gas=179
//   OpStructType: count=1 gas=174
//   OpTypeSwitch: count=4 gas=684
// Store:
//   StoreGetObject: count=1 size=0
// HotSpots:
//   opcodes_types.gno:11 OpTypeSwitch: count=9 gas=518
//   opcodes_types.gno:13 -: count=8 gas=800
//   opcodes_types.gno:27 OpTypes: count=5 gas=398
//   opcodes_types.gno:34 OpTypes: count=2 gas=108
// SubOps:
//   AssignVar: count=1
//   DefineVar: count=2
//   RangeKey: count=4
//   RangeValue: count=4
// Variables:
//   opcodes_types.gno:11 values: count=1
//   opcodes_types.gno:27 t: count=1
//   opcodes_types.gno:34 _: count=1
