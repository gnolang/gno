package vm

import (
	_ "embed"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/jaekwon/testify/assert"
	"sync"
	"testing"
)

var simulator *Simulator
var wg sync.WaitGroup

func init() {
	var err error
	simulator, err = NewSimulator(true, "../../../../gnovm/stdlibs")
	if err != nil {
		panic(err)
	}
	simuAddPkg()
	wg.Add(2)
	go simulator.startServer(&wg)
	go simulator.ibc.OnRecvPacket(&wg)
}

func simuAddPkg() {
	simulator.addPkgFromPath("../../../../examples/gno.land/r/demo/hello/", "gno.land/r/demo/hello")
	simulator.addPkgFromPath("../../../../examples/gno.land/r/demo/greet/", "gno.land/r/demo/greet")
}

//go:embed simulation_data/msg_call_success.json
var msgCallSuccessBz []byte

// No innerMsgs nested
func TestMsgSuccess(t *testing.T) {
	res, _ := simulator.simuCall([][]*std.MemFile{}, msgCallSuccessBz)
	assert.NoError(t, res.Error)
	wg.Wait()
}
