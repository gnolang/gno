package vmk

import (
	_ "embed"
	"sync"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/jaekwon/testify/assert"
)

var simulator *Simulator
var wg sync.WaitGroup

func init() {
	var err error
	simulator, err = NewSimulator(true, "../../gnovm/stdlibs")
	if err != nil {
		panic(err)
	}
	simuAddPkg()
}

func simuAddPkg() {
	simulator.addPkgFromPath("../../examples/gno.land/r/demo/hello/", "gno.land/r/demo/hello")
	simulator.addPkgFromPath("../../examples/gno.land/r/demo/greet/", "gno.land/r/demo/greet")
	simulator.addPkgFromPath("../../examples/gno.land/r/demo/hola/", "gno.land/r/demo/hola")
}

//go:embed simulation_data/msg_call_success.json
var msgCallSuccessBz []byte

// func TestInternalCallSuccess(t *testing.T) {
// 	// bootstrap handleMsg routine
// 	wg := &sync.WaitGroup{}
// 	go simulator.VMKpr.HandleMsg(wg)
// 	wg.Wait()

// 	res, _ := simulator.simuCall([][]*std.MemFile{}, msgCallSuccessBz)
// 	assert.NoError(t, res.Error)
// 	time.Sleep(1 * time.Second)
// }

func TestIBCCallSuccess(t *testing.T) {
	// bootstrap handleMsg routine
	wg := &sync.WaitGroup{}
	go simulator.VMKpr.HandleMsg(wg)

	go simulator.ibc.OnRecvPacket()
	go simulator.ibc.OnAcknowledgementPacket()
	res, _ := simulator.simuCall([][]*std.MemFile{}, msgCallSuccessBz)

	wg.Wait()
	assert.NoError(t, res.Error)
	time.Sleep(1 * time.Second)
}
