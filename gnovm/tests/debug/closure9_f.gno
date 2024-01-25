package main

// 1. we can determine target vars by hasClosure pattern and externNames(without uverse).
// it can be a loopvar, or vars derived from loopvar, e.g. x := i;
// 2. now we need to capture every transient state of the target var; firstly we should
// eval the transient state, there are two ways around:
// a) eval funcLit, but the time doing this is hard to determine, namely, you have to eval
// before it leaves its using context. In some sense, you have to cover all cases when a funcValue
// is used, like assign, xxxLit, funcCall as an arg, etc.
// b) another way for this is to generate a `time series` values that the index is the loop index,
// e.g. in a for loop, we define a new var like x := i, we store the transient state of x per loop
// in the time series, which is a slice. This slice is used when the closure fv is called,
// replacing the var in default block.(func value have a slice of name indicates it's captured, when
// eval name in this slice, using the time series instead).
// this solution mainly differs that it eval target x independently with the closure funcLit, so it avoids
// the obsession to determine the eval order. this seems the right way. hmmm.

//========================================================================================================
// work flow 1.0, hard to do
// 1. determine loop var, whole logic depends on this, the key word is dynamic, it causes polymorphic.
// this should most be done via define(), // TODO: check it
// 2. in transcribe, find if loopvar is used somewhere, to identify pattern like x := i, what to do with
// other use cases? this brings complexity. This is wrong feel.
// we can generate the ts slice for every target var, e.g. x in this case. if two vars, two slices. so fv should
// reference a slice of slice.

// work flow 2.0
// 1. ignore loop var, assume all externNames apart from uverse is target captured var, this also needs to happen
// in transcribe, tag it, and can be checked everywhere it's appears in runtime, x := i, or var x int, etc.
// when eval funcLit, new a slice of slice and push value for further update.

func main() {
	var fns []func() int
	for i := 0; i < 5; i++ {
		//x := i // check if x is polymorphic, if yes, new a slice and set value by index
		var x int // new a slice no init, this would be updated later, so the slice can be mutated after opFuncLit
		f := func() int {
			return x // set the slice to fv in
		}
		x = i                // check if x is polymorphic, is yes, update slice assigned to fv before this, which is a reference.
		fns = append(fns, f) // where should eval funcLit, where it assigned somewhere, or used in another func lit
	}
	for _, fn := range fns {
		println(fn())
	}
}

// Output:
// 0
// 1
// 2
// 3
// 4
