package os

import (
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
)

func PrintAllGoroutines() {
	// XXX is this the best way?
	func() {
		defer func() {
			if r := recover(); r != nil {
				// #1
				buf := make([]byte, 1<<16)
				stackSize := runtime.Stack(buf, true)
				fmt.Printf("%s\n", string(buf[0:stackSize]))

				// #2
				os.Stdout.Write([]byte("pprof.Lookup('goroutine'):\n"))
				pprof.Lookup("goroutine").WriteTo(os.Stdout, 1)
			}
		}()
		panic("THIS_PANIC_INDUCED_FOR_DEBUGGING") // XXX
	}()
}
