package t55runner

import "testing"

var saved func()

func init() { saved = func() { count++ } }

func TestRunnerWrites(t *testing.T) {
	before := count
	count++
	if count != before+1 {
		t.Fatalf("the runner's call of a test must write")
	}
}

func TestHelperRefused(t *testing.T) {
	before := count
	var r any
	func() {
		defer func() { r = recover() }()
		testing.AllocsPerRun2(1, saved)
	}()
	if r == nil || count != before {
		t.Fatalf("a stored value called by a testing helper must not write: %v %d", r, count)
	}
}

func ExampleCount() {
	count = 7
	println(Count())
	// Output:
	// 7
}
