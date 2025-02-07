package fail

import (
	"fmt"
	"os"
	"strconv"
	"sync"
)

func setFromEnv() {
	callIndexToFailS := os.Getenv("FAIL_TEST_INDEX")

	if callIndexToFailS == "" {
		callIndexToFail = -1
	} else {
		var err error
		callIndexToFail, err = strconv.Atoi(callIndexToFailS)
		if err != nil {
			callIndexToFail = -1
		}
	}
}

var (
	callIndex           int       // indexes Fail calls
	callIndexToFail     int       // index of call which should fail
	callIndexToFailOnce sync.Once // sync.Once to set the value of the above
)

// Fail exits the program when after being called the same number of times as
// that passed as the FAIL_TEST_INDEX environment variable.
func Fail() {
	callIndexToFailOnce.Do(setFromEnv)

	if callIndex == callIndexToFail {
		Exit()
	}

	callIndex += 1
}

func Exit() {
	fmt.Printf("*** fail-test %d ***\n", callIndex)
	os.Exit(1)
	//	proc, _ := os.FindProcess(os.Getpid())
	//	proc.Signal(os.Interrupt)
	//	panic(fmt.Sprintf("*** fail-test %d ***", callIndex))
}
